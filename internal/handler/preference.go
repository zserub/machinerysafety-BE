package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// PreferenceResponse is the API response format for preferences.
type PreferenceResponse struct {
	Theme    string          `json:"theme"`    // "dark" or "light"
	Language string          `json:"language"` // "en" or "hu"
	Columns  map[string]bool `json:"columns"`  // column visibility settings
}

// PreferenceUpdateRequest is the request body for updating preferences.
type PreferenceUpdateRequest struct {
	Theme    *string          `json:"theme,omitempty"`    // "dark" or "light"
	Language *string          `json:"language,omitempty"` // "en" or "hu"
	Columns  map[string]bool  `json:"columns,omitempty"`  // column visibility settings
}

// PreferenceHandler handles preference endpoints.
type PreferenceHandler struct {
	preferenceRepo repo.PreferenceRepository
	logger         *slog.Logger
}

// NewPreferenceHandler creates a new preference handler.
func NewPreferenceHandler(preferenceRepo repo.PreferenceRepository, logger *slog.Logger) *PreferenceHandler {
	return &PreferenceHandler{
		preferenceRepo: preferenceRepo,
		logger:         logger.With("component", "handler.preference"),
	}
}

// Get handles GET /api/user/preferences
func (h *PreferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pref, err := h.preferenceRepo.GetByPersonID(r.Context(), person.PersonID)
	if err != nil {
		h.logger.Error("failed to fetch preferences", "person_id", person.PersonID, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, toPreferenceResponse(pref))
}

// Update handles PATCH /api/user/preferences
func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req PreferenceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("invalid request body", "error", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate and convert request
	updates, err := h.validateAndConvert(&req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	pref, err := h.preferenceRepo.Update(r.Context(), person.PersonID, updates)
	if err != nil {
		h.logger.Error("failed to update preferences", "person_id", person.PersonID, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, toPreferenceResponse(pref))
}

// validateAndConvert validates and converts the request to domain update.
func (h *PreferenceHandler) validateAndConvert(req *PreferenceUpdateRequest) (*domain.PreferenceUpdate, error) {
	updates := &domain.PreferenceUpdate{}

	if req.Theme != nil {
		switch *req.Theme {
		case "dark":
			theme := false
			updates.Theme = &theme
		case "light":
			theme := true
			updates.Theme = &theme
		default:
			return nil, &ValidationError{Field: "theme", Message: "must be 'dark' or 'light'"}
		}
	}

	if req.Language != nil {
		switch *req.Language {
		case "en", "hu":
			updates.Language = req.Language
		default:
			return nil, &ValidationError{Field: "language", Message: "must be 'en' or 'hu'"}
		}
	}

	if req.Columns != nil {
		updates.Columns = req.Columns
	}

	return updates, nil
}

// toPreferenceResponse converts domain preference to API response.
func toPreferenceResponse(pref *domain.Preference) PreferenceResponse {
	theme := "dark"
	if pref.Theme {
		theme = "light"
	}

	// Convert ActiveColumns to map[string]bool
	columns := make(map[string]bool)
	if pref.ActiveColumns != nil {
		if colMap, ok := pref.ActiveColumns.(map[string]interface{}); ok {
			for k, v := range colMap {
				if boolVal, ok := v.(bool); ok {
					columns[k] = boolVal
				}
			}
		}
	}

	return PreferenceResponse{
		Theme:    theme,
		Language: pref.Language,
		Columns:  columns,
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
