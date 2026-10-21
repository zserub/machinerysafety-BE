package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/service"
)

// CommentHandler handles comment endpoints.
type CommentHandler struct {
	commentService *service.CommentService
	logger         *slog.Logger
}

// NewCommentHandler creates a new comment handler.
func NewCommentHandler(commentService *service.CommentService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		logger:         logger.With("component", "handler.comment"),
	}
}

// CreateCommentRequest is the request body for creating a comment.
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// CommentResponse is the JSON response for a comment.
type CommentResponse struct {
	ID         int64   `json:"id"`
	RiskID     int64   `json:"riskId"`
	AuthorID   int16   `json:"authorId"`
	AuthorName string  `json:"authorName"`
	Content    string  `json:"content"`
	CreatedAt  string  `json:"createdAt"`
	EditedAt   *string `json:"editedAt,omitempty"`
}

// toCommentResponse converts a domain.Comment to CommentResponse.
func toCommentResponse(c *domain.Comment) CommentResponse {
	resp := CommentResponse{
		ID:         c.ID,
		RiskID:     c.RiskID,
		AuthorID:   c.AuthorID,
		AuthorName: c.AuthorName,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.EditedAt != nil {
		edited := c.EditedAt.Format("2006-01-02T15:04:05Z")
		resp.EditedAt = &edited
	}
	return resp
}

// toCommentResponses converts a slice of domain.Comment to []CommentResponse.
func toCommentResponses(comments []domain.Comment) []CommentResponse {
	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = toCommentResponse(&c)
	}
	return responses
}

// List handles GET /api/entries/{riskId}/comments
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	riskIDStr := chi.URLParam(r, "riskId")
	riskID, err := strconv.ParseInt(riskIDStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid riskId")
		return
	}

	comments, err := h.commentService.List(r.Context(), riskID)
	if err != nil {
		if err == domain.ErrRiskNotFound {
			Error(w, http.StatusNotFound, "risk not found")
			return
		}
		h.logger.Error("failed to list comments", "risk_id", riskID, "error", err)
		ErrorWithDetails(w, http.StatusInternalServerError, "failed to list comments", err.Error())
		return
	}

	JSON(w, http.StatusOK, toCommentResponses(comments))
}

// Create handles POST /api/entries/{riskId}/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	riskIDStr := chi.URLParam(r, "riskId")
	riskID, err := strconv.ParseInt(riskIDStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid riskId")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("create: invalid request body", "error", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate content
	if req.Content == "" {
		Error(w, http.StatusBadRequest, "content is required")
		return
	}
	if len(req.Content) > 500 {
		Error(w, http.StatusBadRequest, "content must be at most 500 characters")
		return
	}

	input := service.CreateCommentInput{
		RiskID:  riskID,
		Content: req.Content,
	}

	comment, err := h.commentService.Create(r.Context(), input, person.PersonID)
	if err != nil {
		if err == domain.ErrRiskNotFound {
			Error(w, http.StatusNotFound, "risk not found")
			return
		}
		h.logger.Error("failed to create comment", "risk_id", riskID, "error", err)
		ErrorWithDetails(w, http.StatusInternalServerError, "failed to create comment", err.Error())
		return
	}

	h.logger.Info("comment created", "id", comment.ID, "risk_id", riskID, "author_id", person.PersonID)
	JSON(w, http.StatusCreated, toCommentResponse(comment))
}
