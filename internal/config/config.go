package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration.
type Config struct {
	// Server
	ServerHost         string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerPort         int           `env:"SERVER_PORT" envDefault:"8080"`
	ServerReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
	ServerWriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	ServerIdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`

	// Database (PostgreSQL)
	DBHost            string        `env:"DB_HOST" envDefault:"localhost"`
	DBPort            int           `env:"DB_PORT" envDefault:"5432"`
	DBName            string        `env:"DB_NAME,required"`
	DBUser            string        `env:"DB_USER,required"`
	DBPassword        string        `env:"DB_PASSWORD,required"`
	DBSSLMode         string        `env:"DB_SSL_MODE" envDefault:"disable"`
	DBSchema          string        `env:"DB_SCHEMA" envDefault:"public"`
	DBMaxConns        int32         `env:"DB_MAX_CONNS" envDefault:"25"`
	DBMinConns        int32         `env:"DB_MIN_CONNS" envDefault:"5"`
	DBMaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"1h"`
	DBMaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"30m"`

	// MinIO
	MinIOEndpoint        string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKeyID     string `env:"MINIO_ACCESS_KEY_ID,required"`
	MinIOSecretAccessKey string `env:"MINIO_SECRET_ACCESS_KEY,required"`
	MinIOBucket          string `env:"MINIO_BUCKET" envDefault:"photos"`
	MinIOUseSSL          bool   `env:"MINIO_USE_SSL" envDefault:"false"`

	// Session
	SessionKey      string `env:"SESSION_KEY,required"` // 32 or 64 bytes for AES-256
	SessionName     string `env:"SESSION_NAME" envDefault:"machinery_session"`
	SessionMaxAge   int    `env:"SESSION_MAX_AGE" envDefault:"86400"` // 24 hours
	SessionSecure   bool   `env:"SESSION_SECURE" envDefault:"false"`  // true for HTTPS
	SessionHTTPOnly bool   `env:"SESSION_HTTP_ONLY" envDefault:"true"`
	SessionSameSite string `env:"SESSION_SAME_SITE" envDefault:"lax"` // strict, lax, none

	// Logging
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`  // debug, info, warn, error
	LogFormat   string `env:"LOG_FORMAT" envDefault:"json"` // json, text
	LogFilePath string `env:"LOG_FILE_PATH"`                // optional: path to log file

	// Migrations
	RunMigrations  bool   `env:"RUN_MIGRATIONS" envDefault:"false"`
	MigrationsPath string `env:"MIGRATIONS_PATH" envDefault:"file://migrations"`
	SeedsPath      string `env:"SEEDS_PATH"` // dev only: mock data applied after migrations (e.g. file://seeds)

	// Initial admin: created on startup only if the persons table is empty
	InitialAdminUsername string `env:"INITIAL_ADMIN_USERNAME"`
	InitialAdminPassword string `env:"INITIAL_ADMIN_PASSWORD"`

	// CORS
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`

	// Bcrypt
	BcryptCost int `env:"BCRYPT_COST" envDefault:"10"` // 10-14 recommended

	// Metabase
	MetabaseURL        string `env:"METABASE_URL" envDefault:"http://localhost:3000"`
	MetabasePublicURL  string `env:"METABASE_PUBLIC_URL"` // Browser-accessible URL; defaults to MetabaseURL
	MetabaseAPIKey     string `env:"METABASE_API_KEY"`
	MetabaseSecretKey  string `env:"METABASE_SECRET_KEY"` // For JWT signed embeds
	RunMetabaseSetup   bool   `env:"RUN_METABASE_SETUP" envDefault:"false"`
	MetabaseDatabaseID int    `env:"METABASE_DATABASE_ID" envDefault:"3"`
}

// MetabaseExternalURL returns the browser-accessible Metabase URL.
// Falls back to MetabaseURL if MetabasePublicURL is not set.
func (c *Config) MetabaseExternalURL() string {
	if c.MetabasePublicURL != "" {
		return c.MetabasePublicURL
	}
	return c.MetabaseURL
}

// Load parses environment variables into Config struct.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// MustLoad calls Load and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

// DatabaseURL returns the PostgreSQL connection string.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode, c.DBSchema,
	)
}

// ServerAddress returns the formatted server address.
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}
