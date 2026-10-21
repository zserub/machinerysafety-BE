package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/config"
	"github.com/zserub/machinerysafety-BE/internal/filestore"
	"github.com/zserub/machinerysafety-BE/internal/handler"
	"github.com/zserub/machinerysafety-BE/internal/metabase"
	"github.com/zserub/machinerysafety-BE/internal/repo"
	"github.com/zserub/machinerysafety-BE/internal/service"
	"github.com/zserub/machinerysafety-BE/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg)
	logger.Info("starting machinery safety backend",
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
	)

	// Create root context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run database migrations (optional)
	if cfg.RunMigrations {
		if err := runMigrations(cfg, logger); err != nil {
			logger.Error("migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("skipping migrations", "run_migrations", cfg.RunMigrations)
	}

	// Initialize database pool
	pool, err := initDB(ctx, cfg, logger)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Initialize repos
	personRepo := repo.NewPersonRepo(pool, logger)

	// Create the first admin on a fresh install (no-op once any user exists)
	if err := ensureInitialAdmin(ctx, cfg, personRepo, logger); err != nil {
		logger.Error("initial admin creation failed", "error", err)
		os.Exit(1)
	}

	// Initialize MinIO
	minioStore, err := filestore.NewMinioStore(cfg, logger)
	if err != nil {
		logger.Error("MinIO initialization failed", "error", err)
		os.Exit(1)
	}

	// Ensure bucket exists
	if err := minioStore.EnsureBucket(ctx); err != nil {
		logger.Error("failed to ensure MinIO bucket", "error", err)
		os.Exit(1)
	}

	// Initialize Metabase client
	metabaseClient := metabase.NewClient(cfg, logger)

	// Run Metabase dashboard setup (optional)
	if cfg.RunMetabaseSetup {
		if cfg.MetabaseAPIKey == "" {
			logger.Warn("skipping metabase setup - no API key configured")
		} else {
			setupService := metabase.NewSetupService(metabaseClient, cfg.MetabaseDatabaseID, logger)
			result, err := setupService.Run(ctx)
			if err != nil {
				logger.Error("metabase setup failed", "error", err)
				os.Exit(1)
			}
			if result.Skipped {
				logger.Info("metabase setup skipped - dashboard already exists",
					"dashboard_id", result.DashboardID)
			} else {
				logger.Info("metabase setup completed",
					"dashboard_id", result.DashboardID,
					"cards_created", result.CardsCreated,
					"cards_skipped", result.CardsSkipped)
			}
		}
	}

	metadataRepo := repo.NewMetadataRepo(pool, logger)
	riskRepo := repo.NewRiskRepo(pool, logger)
	preferenceRepo := repo.NewPreferenceRepo(pool, logger)
	commentRepo := repo.NewCommentRepo(pool, logger)

	// Initialize services
	authService := service.NewAuthService(personRepo, logger)
	riskService := service.NewRiskService(riskRepo, logger)
	commentService := service.NewCommentService(commentRepo, riskRepo, logger)

	// Initialize session manager
	sessionMgr := handler.NewSessionManager(cfg)

	// Initialize middleware
	authMiddleware := handler.NewAuthMiddleware(sessionMgr, personRepo, logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(pool, minioStore)
	authHandler := handler.NewAuthHandler(authService, sessionMgr, logger)
	metadataHandler := handler.NewMetadataHandler(metadataRepo, logger)
	riskHandler := handler.NewRiskHandler(riskService, metadataRepo, logger)
	preferenceHandler := handler.NewPreferenceHandler(preferenceRepo, logger)
	commentHandler := handler.NewCommentHandler(commentService, logger)
	metabaseHandler := handler.NewMetabaseHandler(metabaseClient, cfg.MetabaseExternalURL(), logger)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Routes
	r.Get("/health", healthHandler.Health)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)

	// Version endpoint
	r.Get("/api/version", handler.VersionHandler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.With(authMiddleware.RequireAuth).Get("/me", authHandler.Me)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// User endpoints
			r.Get("/user/permissions", authHandler.UserPermissions)
			r.Get("/user/preferences", preferenceHandler.Get)
			r.Patch("/user/preferences", preferenceHandler.Update)

			// Metadata endpoint
			r.Get("/metadata", metadataHandler.GetAll)

			// Risk CRUD endpoints
			r.Route("/entries", func(r chi.Router) {
				r.With(authMiddleware.RequirePermission("entries", "view")).Get("/", riskHandler.List)
				r.With(authMiddleware.RequirePermission("entries", "view")).Get("/{id}", riskHandler.GetByID)
				r.With(authMiddleware.RequirePermission("entries", "create")).Post("/", riskHandler.Create)
				r.With(authMiddleware.RequirePermission("entries", "edit")).Patch("/{id}", riskHandler.Update)
				r.With(authMiddleware.RequirePermission("entries", "delete")).Delete("/{id}", riskHandler.Delete)

				// Comment sub-routes
				r.Route("/{riskId}/comments", func(r chi.Router) {
					r.With(authMiddleware.RequirePermission("entries", "view")).Get("/", commentHandler.List)
					r.With(authMiddleware.RequirePermission("entries", "create")).Post("/", commentHandler.Create)
				})
			})

			// Metabase endpoints
			r.Route("/metabase", func(r chi.Router) {
				r.With(authMiddleware.RequirePermission("statistics", "view")).Get("/config", metabaseHandler.Config)
				r.With(authMiddleware.RequirePermission("statistics", "view")).Get("/dashboards/{id}/embed", metabaseHandler.DashboardEmbed)
				r.With(authMiddleware.RequirePermission("statistics", "view")).Get("/cards/{id}/embed", metabaseHandler.CardEmbed)
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", "address", cfg.ServerAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped")
}

func initLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var writer io.Writer = os.Stdout

	if cfg.LogFilePath != "" {
		file, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file %s: %v\n", cfg.LogFilePath, err)
		} else {
			writer = io.MultiWriter(os.Stdout, file)
		}
	}

	var h slog.Handler
	if strings.ToLower(cfg.LogFormat) == "text" {
		h = slog.NewTextHandler(writer, opts)
	} else {
		h = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(h)
	slog.SetDefault(logger)
	return logger
}

func initDB(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	logger.Info("connecting to database",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = cfg.DBMaxConns
	poolConfig.MinConns = cfg.DBMinConns
	poolConfig.MaxConnLifetime = cfg.DBMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DBMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established")
	return pool, nil
}

func runMigrations(cfg *config.Config, logger *slog.Logger) error {
	logger.Info("running database migrations", "path", cfg.MigrationsPath)

	m, err := migrate.New(cfg.MigrationsPath, cfg.DatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	logger.Info("migrations completed", "version", version, "dirty", dirty)

	if cfg.SeedsPath == "" {
		return nil
	}

	// Dev seeds are versioned in their own table, so they never mix with schema versions
	logger.Warn("applying dev seed data", "path", cfg.SeedsPath)
	seeder, err := migrate.New(cfg.SeedsPath, cfg.DatabaseURL()+"&x-migrations-table=schema_migrations_seeds")
	if err != nil {
		return fmt.Errorf("failed to create seeder: %w", err)
	}
	defer seeder.Close()

	if err := seeder.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply seeds: %w", err)
	}
	return nil
}

const minAdminPasswordLength = 8

// ensureInitialAdmin creates an admin from INITIAL_ADMIN_* when the database has no users yet.
func ensureInitialAdmin(ctx context.Context, cfg *config.Config, personRepo *repo.PersonRepo, logger *slog.Logger) error {
	if cfg.InitialAdminUsername == "" || cfg.InitialAdminPassword == "" {
		return nil
	}
	if len(cfg.InitialAdminPassword) < minAdminPasswordLength {
		return fmt.Errorf("INITIAL_ADMIN_PASSWORD must be at least %d characters", minAdminPasswordLength)
	}

	hash, err := utils.HashPassword(cfg.InitialAdminPassword, cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	created, err := personRepo.CreateInitialAdmin(ctx, cfg.InitialAdminUsername, hash)
	if err != nil {
		return err
	}
	if created {
		logger.Warn("created initial admin user - remove INITIAL_ADMIN_PASSWORD from the environment",
			"username", cfg.InitialAdminUsername)
	}
	return nil
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration_ms", time.Since(start).Milliseconds(),
					"bytes", ww.BytesWritten(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
