package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger interface for services that can be pinged.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db        *pgxpool.Pool
	filestore Pinger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *pgxpool.Pool, filestore Pinger) *HealthHandler {
	return &HealthHandler{
		db:        db,
		filestore: filestore,
	}
}

// HealthResponse is the response format for health checks.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Health returns overall health status.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]string),
	}

	// Check database
	if err := h.db.Ping(ctx); err != nil {
		response.Services["database"] = "unhealthy"
		response.Status = "unhealthy"
	} else {
		response.Services["database"] = "healthy"
	}

	// Check MinIO
	if err := h.filestore.Ping(ctx); err != nil {
		response.Services["filestore"] = "unhealthy"
		response.Status = "unhealthy"
	} else {
		response.Services["filestore"] = "healthy"
	}

	status := http.StatusOK
	if response.Status == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	JSON(w, status, response)
}

// Liveness indicates the server is running.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

// Readiness indicates the server can accept traffic.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	h.Health(w, r)
}

// Version is set at build time via ldflags.
var Version = "dev"

// VersionResponse is the response format for the version endpoint.
type VersionResponse struct {
	Version string `json:"version"`
}

// VersionHandler returns the API version.
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, VersionResponse{Version: Version})
}
