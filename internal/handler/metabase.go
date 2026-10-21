package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zserub/machinerysafety-BE/internal/metabase"
)

// MetabaseHandler handles Metabase embed endpoints.
type MetabaseHandler struct {
	client    *metabase.Client
	publicURL string // browser-accessible Metabase URL
	logger    *slog.Logger
}

// NewMetabaseHandler creates a new MetabaseHandler.
func NewMetabaseHandler(client *metabase.Client, publicURL string, logger *slog.Logger) *MetabaseHandler {
	return &MetabaseHandler{
		client:    client,
		publicURL: publicURL,
		logger:    logger.With("component", "handler.metabase"),
	}
}

// MetabaseConfigResponse is returned by the Config endpoint.
type MetabaseConfigResponse struct {
	Configured  bool   `json:"configured"`
	EmbedURL    string `json:"embed_url,omitempty"`
	MetabaseURL string `json:"metabase_url,omitempty"`
}

// Config returns the Metabase dashboard embed URL for the frontend.
// Finds the analytics dashboard by name, creates a public link if needed.
func (h *MetabaseHandler) Config(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary, err := h.client.FindDashboardByName(ctx, metabase.DashboardName)
	if err != nil || summary == nil {
		JSON(w, http.StatusOK, MetabaseConfigResponse{Configured: false})
		return
	}

	dashboard, err := h.client.GetDashboard(ctx, summary.ID)
	if err != nil {
		JSON(w, http.StatusOK, MetabaseConfigResponse{Configured: false})
		return
	}

	uuid := dashboard.PublicUUID
	if uuid == "" {
		uuid, err = h.client.CreateDashboardPublicLink(ctx, dashboard.ID)
		if err != nil {
			JSON(w, http.StatusOK, MetabaseConfigResponse{Configured: false})
			return
		}
	}

	embedURL := h.publicURL + "/public/dashboard/" + uuid
	JSON(w, http.StatusOK, MetabaseConfigResponse{
		Configured:  true,
		EmbedURL:    embedURL,
		MetabaseURL: h.publicURL,
	})
}

// EmbedResponse is the response for embed endpoints.
type EmbedResponse struct {
	EmbedURL   string `json:"embed_url"`
	IframeHTML string `json:"iframe_html"`
}

// DashboardEmbed returns the public embed URL for a dashboard.
func (h *MetabaseHandler) DashboardEmbed(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}

	ctx := r.Context()

	// Fetch dashboard to get public UUID (or check if already shared)
	dashboard, err := h.client.GetDashboard(ctx, id)
	if err != nil {
		h.handleMetabaseError(w, err, "dashboard", id)
		return
	}

	// If dashboard doesn't have a public UUID, create one
	uuid := dashboard.PublicUUID
	if uuid == "" {
		uuid, err = h.client.CreateDashboardPublicLink(ctx, id)
		if err != nil {
			h.handleMetabaseError(w, err, "dashboard", id)
			return
		}
	}

	embedURL := h.client.GetPublicDashboardURL(uuid)
	iframeHTML := generateIframeHTML(embedURL, dashboard.Name)

	h.logger.Info("dashboard embed URL generated",
		"dashboard_id", id,
		"name", dashboard.Name,
	)

	JSON(w, http.StatusOK, EmbedResponse{
		EmbedURL:   embedURL,
		IframeHTML: iframeHTML,
	})
}

// CardEmbed returns the public embed URL for a card (question).
func (h *MetabaseHandler) CardEmbed(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid card id")
		return
	}

	ctx := r.Context()

	// Fetch card to get public UUID (or check if already shared)
	card, err := h.client.GetCard(ctx, id)
	if err != nil {
		h.handleMetabaseError(w, err, "card", id)
		return
	}

	// If card doesn't have a public UUID, create one
	uuid := card.PublicUUID
	if uuid == "" {
		uuid, err = h.client.CreateCardPublicLink(ctx, id)
		if err != nil {
			h.handleMetabaseError(w, err, "card", id)
			return
		}
	}

	embedURL := h.client.GetPublicCardURL(uuid)
	iframeHTML := generateIframeHTML(embedURL, card.Name)

	h.logger.Info("card embed URL generated",
		"card_id", id,
		"name", card.Name,
	)

	JSON(w, http.StatusOK, EmbedResponse{
		EmbedURL:   embedURL,
		IframeHTML: iframeHTML,
	})
}

// handleMetabaseError converts Metabase errors to HTTP responses.
func (h *MetabaseHandler) handleMetabaseError(w http.ResponseWriter, err error, resourceType string, id int) {
	h.logger.Error("metabase error",
		"resource_type", resourceType,
		"id", id,
		"error", err,
	)

	if errors.Is(err, metabase.ErrNotFound) {
		Error(w, http.StatusNotFound, resourceType+" not found")
		return
	}
	if errors.Is(err, metabase.ErrUnauthorized) {
		Error(w, http.StatusInternalServerError, "metabase authentication failed")
		return
	}
	if errors.Is(err, metabase.ErrPublicSharingOff) {
		Error(w, http.StatusBadRequest, "public sharing is disabled in metabase")
		return
	}

	Error(w, http.StatusInternalServerError, "metabase error")
}

// generateIframeHTML creates an iframe HTML snippet for embedding.
func generateIframeHTML(embedURL, title string) string {
	return `<iframe src="` + embedURL + `" title="` + title + `" frameborder="0" width="100%" height="600" allowtransparency></iframe>`
}
