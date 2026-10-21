package handler

import (
	"log/slog"
	"net/http"

	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// MetadataHandler handles metadata endpoints.
type MetadataHandler struct {
	metadataRepo *repo.MetadataRepo
	logger       *slog.Logger
}

// NewMetadataHandler creates a new metadata handler.
func NewMetadataHandler(metadataRepo *repo.MetadataRepo, logger *slog.Logger) *MetadataHandler {
	return &MetadataHandler{
		metadataRepo: metadataRepo,
		logger:       logger.With("component", "handler.metadata"),
	}
}

// GetAll handles GET /api/v1/metadata
func (h *MetadataHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metadata, err := h.metadataRepo.GetAll(r.Context())
	if err != nil {
		h.logger.Error("failed to fetch metadata", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, metadata)
}
