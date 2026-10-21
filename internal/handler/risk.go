package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
	"github.com/zserub/machinerysafety-BE/internal/service"
)

// RiskHandler handles risk endpoints.
type RiskHandler struct {
	riskService  *service.RiskService
	metadataRepo *repo.MetadataRepo
	logger       *slog.Logger
}

// RiskResponse is the JSON response format for a risk entry.
// Uses string values only for lookup fields (no redundant ID fields).
type RiskResponse struct {
	ID                 int64                     `json:"ID"`
	RiskType           int                       `json:"RiskType"`
	LinkedRiskID       *int64                    `json:"LinkedRiskID"`
	CreationDate       time.Time                 `json:"CreationDate"`
	Machines           string                    `json:"Machines"`
	DangerDescription  string                    `json:"DangerDescription"`
	Action             *string                   `json:"Action"`
	MOCNumber          *int                      `json:"MOCNumber"`
	Deadline           *time.Time                `json:"Deadline"`
	OriginalDeadline   *time.Time                `json:"OriginalDeadline"`
	ImplementationDate *time.Time                `json:"ImplementationDate"`
	EntryModifiedAt    time.Time                 `json:"EntryModifiedAt"`
	CommittedAt        *time.Time                `json:"CommittedAt"`
	Cost               *float64                  `json:"Cost"`
	ExpenseType        *string                   `json:"ExpenseType"`
	Line               string                    `json:"Line"`
	Category           *string                   `json:"Category"`
	RiskLevel          string                    `json:"RiskLevel"`
	State              string                    `json:"State"`
	Responsible        *string                   `json:"Responsible"`
	CreatedBy          *string                   `json:"CreatedBy"`
	LinkedRisk         *LinkedRiskResponse       `json:"LinkedRisk"`
	Photos             []PhotoResponse           `json:"Photos"`
	StatusHistory      []StatusHistoryResponse   `json:"StatusHistory"`
	DeadlineHistory    []DeadlineHistoryResponse `json:"DeadlineHistory"`
	Postponements      int                       `json:"Postponements"`
}

// LinkedRiskResponse is the JSON response for a linked risk summary.
type LinkedRiskResponse struct {
	ID                int64  `json:"ID"`
	Machines          string `json:"Machines"`
	Line              string `json:"Line"`
	DangerDescription string `json:"DangerDescription"`
	RiskType          int    `json:"RiskType"`
}

// PhotoResponse is the JSON response for a photo.
type PhotoResponse struct {
	RiskID      int64     `json:"RiskID"`
	PhotoNumber int       `json:"PhotoNumber"`
	FilePath    string    `json:"FilePath"`
	UploadedAt  time.Time `json:"UploadedAt"`
}

// StatusHistoryResponse is the JSON response for a status history entry.
type StatusHistoryResponse struct {
	ID        int64     `json:"ID"`
	Timestamp time.Time `json:"Timestamp"`
	User      string    `json:"User"`
	Action    string    `json:"Action"`
}

// DeadlineHistoryResponse is the JSON response for a deadline history entry.
type DeadlineHistoryResponse struct {
	ID        int64     `json:"ID"`
	Timestamp time.Time `json:"Timestamp"`
	User      string    `json:"User"`
	Action    string    `json:"Action"`
	Details   string    `json:"Details"`
	Reason    string    `json:"Reason,omitempty"`
}

// RiskListResponse is the JSON response for a list of risks.
type RiskListResponse struct {
	Entries    []RiskResponse `json:"entries"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}

// toRiskResponse converts a domain.Risk to a RiskResponse.
// includeHistory controls whether to include status/deadline history (only for detail view).
func toRiskResponse(r *domain.Risk, includeHistory bool) RiskResponse {
	resp := RiskResponse{
		ID:                 r.ID,
		RiskType:           r.RiskType,
		LinkedRiskID:       r.LinkedRiskID,
		CreationDate:       r.CreationDate,
		Machines:           r.Machines,
		DangerDescription:  r.DangerDescription,
		Action:             r.Action,
		MOCNumber:          r.MOCNumber,
		Deadline:           r.Deadline,
		OriginalDeadline:   r.Deadline,
		ImplementationDate: r.ImplementationDate,
		EntryModifiedAt:    r.EntryModifiedAt,
		CommittedAt:        r.CommittedAt,
		Cost:               r.Cost,
		ExpenseType:        r.ExpenseTypeName,
		Line:               r.LineName,
		Category:           r.CategoryName,
		RiskLevel:          r.RiskLevelName,
		State:              r.StateName,
		Responsible:        r.ResponsibleName,
		CreatedBy:          r.CreatedByName,
	}

	// Convert linked risk
	if r.LinkedRisk != nil {
		resp.LinkedRisk = &LinkedRiskResponse{
			ID:                r.LinkedRisk.ID,
			Machines:          r.LinkedRisk.Machines,
			Line:              r.LinkedRisk.LineName,
			DangerDescription: r.LinkedRisk.DangerDescription,
			RiskType:          r.LinkedRisk.RiskType,
		}
	}

	// Convert photos
	if len(r.Photos) > 0 {
		resp.Photos = make([]PhotoResponse, len(r.Photos))
		for i, p := range r.Photos {
			resp.Photos[i] = PhotoResponse{
				RiskID:      p.RiskID,
				PhotoNumber: p.PhotoNumber,
				FilePath:    p.FilePath,
				UploadedAt:  p.UploadedAt,
			}
		}
	}

	// Only include history for detail view (not list)
	if includeHistory {
		// Convert status history - entries are ordered DESC by changed_at
		// Add synthesized "created" entry at the end (oldest first when reversed)
		var statusHistoryItems []StatusHistoryResponse
		for _, h := range r.StatusHistory {
			statusHistoryItems = append(statusHistoryItems, StatusHistoryResponse{
				ID:        h.ID,
				Timestamp: h.ChangedAt,
				User:      h.ChangedByName,
				Action:    statusChangeToAction(h.NewStateName),
			})
		}
		// Add synthesized "created" entry (uses creation date and created_by from risk)
		createdByName := ""
		if r.CreatedByName != nil {
			createdByName = *r.CreatedByName
		}
		statusHistoryItems = append(statusHistoryItems, StatusHistoryResponse{
			ID:        0,
			Timestamp: r.CreationDate,
			User:      createdByName,
			Action:    "created",
		})
		resp.StatusHistory = statusHistoryItems

		// Convert deadline history
		if len(r.DeadlineHistory) > 0 {
			// History is ordered DESC by changed_at, so the last item's
			// OldDeadline is the deadline before any changes were made.
			original := r.DeadlineHistory[len(r.DeadlineHistory)-1].OldDeadline
			resp.OriginalDeadline = &original

			resp.DeadlineHistory = make([]DeadlineHistoryResponse, len(r.DeadlineHistory))
			for i := range r.DeadlineHistory {
				h := &r.DeadlineHistory[i]
				action := "postponed"
				if h.NewDeadline.Before(h.OldDeadline) {
					action = "advanced"
				} else {
					resp.Postponements++
				}
				resp.DeadlineHistory[i] = DeadlineHistoryResponse{
					ID:        h.ID,
					Timestamp: h.ChangedAt,
					User:      h.ChangedByName,
					Action:    action,
					Details:   fmt.Sprintf("from %s to %s", h.OldDeadline.Format("2006-01-02"), h.NewDeadline.Format("2006-01-02")),
					Reason:    h.Reason,
				}
			}
		}
	}

	return resp
}

// toRiskListResponse converts a repo.RiskListResult to a RiskListResponse.
func toRiskListResponse(result *repo.RiskListResult) RiskListResponse {
	entries := make([]RiskResponse, len(result.Risks))
	for i := range result.Risks {
		entries[i] = toRiskResponse(&result.Risks[i], false) // no history in list
	}
	return RiskListResponse{
		Entries:    entries,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}
}

// NewRiskHandler creates a new risk handler.
func NewRiskHandler(riskService *service.RiskService, metadataRepo *repo.MetadataRepo, logger *slog.Logger) *RiskHandler {
	return &RiskHandler{
		riskService:  riskService,
		metadataRepo: metadataRepo,
		logger:       logger.With("component", "handler.risk"),
	}
}

// CreateRiskRequest is the request body for creating a risk.
// Field names match frontend conventions (string values for lookups).
type CreateRiskRequest struct {
	RiskType          int      `json:"riskType"`
	LinkedRiskID      *int64   `json:"linkedRiskId"`
	Machine           string   `json:"machine"`
	Line              string   `json:"line"`              // e.g., "Production Line A"
	Category          *string  `json:"category"`          // e.g., "mechanical"
	DescriptionOfDanger string `json:"descriptionOfDanger"`
	RiskLevel         string   `json:"riskLevel"`         // e.g., "high"
	Action            *string  `json:"action"`
	Responsible       *string  `json:"responsible"`       // e.g., "John Smith"
	MOC               *string  `json:"moc"`
	Deadline          *string  `json:"deadline"`          // YYYY-MM-DD
	Cost              *float64 `json:"cost"`
	ExpenseType       *string  `json:"expenseType"`       // e.g., "capex"
}

// UpdateRiskRequest is the request body for updating a risk.
// Field names match frontend conventions (string values for lookups).
type UpdateRiskRequest struct {
	Machine            *string  `json:"machine"`
	Line               *string  `json:"line"`              // e.g., "Production Line A"
	Category           *string  `json:"category"`          // e.g., "mechanical"
	DescriptionOfDanger *string `json:"descriptionOfDanger"`
	RiskLevel          *string  `json:"riskLevel"`         // e.g., "high"
	Action             *string  `json:"action"`
	Responsible        *string  `json:"responsible"`       // e.g., "John Smith"
	MOC                *string  `json:"moc"`
	Deadline           *string  `json:"deadline"`
	Status             *string  `json:"status"`            // e.g., "OPEN", "IN_PROGRESS"
	ImplementationDate *string  `json:"implementationDate"`
	Cost               *float64 `json:"cost"`
	ExpenseType        *string  `json:"expenseType"`       // e.g., "capex"
	LinkedRiskID       *int64   `json:"linkedRiskId"`

	StatusChangeReason   *string `json:"statusChangeReason"`
	DeadlineChangeReason *string `json:"deadlineChangeReason"`
}

// List handles GET /api/v1/entries
func (h *RiskHandler) List(w http.ResponseWriter, r *http.Request) {
	// PageSize=0 means return all entries (for client-side pagination)
	filters := repo.RiskFilters{
		Search:    r.URL.Query().Get("search"),
		Page:      parseIntOrDefault(r.URL.Query().Get("page"), 1),
		PageSize:  parseIntOrDefault(r.URL.Query().Get("pageSize"), 0),
		SortBy:    r.URL.Query().Get("sortBy"),
		SortOrder: r.URL.Query().Get("sortOrder"),
	}

	// Parse optional filters
	if rt := r.URL.Query().Get("riskType"); rt != "" {
		if v, err := strconv.Atoi(rt); err == nil {
			filters.RiskType = &v
		}
	}
	if lid := r.URL.Query().Get("lineId"); lid != "" {
		if v, err := strconv.ParseInt(lid, 10, 16); err == nil {
			v16 := int16(v)
			filters.LineID = &v16
		}
	}
	if cid := r.URL.Query().Get("categoryId"); cid != "" {
		if v, err := strconv.Atoi(cid); err == nil {
			filters.CategoryID = &v
		}
	}
	if rlid := r.URL.Query().Get("riskLevelId"); rlid != "" {
		if v, err := strconv.ParseInt(rlid, 10, 16); err == nil {
			v16 := int16(v)
			filters.RiskLevelID = &v16
		}
	}
	if sid := r.URL.Query().Get("statusId"); sid != "" {
		if v, err := strconv.ParseInt(sid, 10, 16); err == nil {
			v16 := int16(v)
			filters.StatusID = &v16
		}
	}
	if rid := r.URL.Query().Get("responsibleId"); rid != "" {
		if v, err := strconv.ParseInt(rid, 10, 16); err == nil {
			v16 := int16(v)
			filters.ResponsibleID = &v16
		}
	}

	result, err := h.riskService.List(r.Context(), filters)
	if err != nil {
		h.logger.Error("failed to list risks", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, toRiskListResponse(result))
}

// GetByID handles GET /api/v1/entries/{id}
func (h *RiskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	risk, err := h.riskService.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrRiskNotFound {
			Error(w, http.StatusNotFound, "risk not found")
			return
		}
		h.logger.Error("failed to get risk", "id", id, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, toRiskResponse(risk, true))
}

// Create handles POST /api/v1/entries
func (h *RiskHandler) Create(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("create: invalid request body", "error", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.RiskType != 1 && req.RiskType != 2 {
		Error(w, http.StatusBadRequest, "riskType must be 1 or 2")
		return
	}
	if req.Machine == "" {
		Error(w, http.StatusBadRequest, "machine is required")
		return
	}
	if req.DescriptionOfDanger == "" {
		Error(w, http.StatusBadRequest, "descriptionOfDanger is required")
		return
	}
	if req.Line == "" {
		Error(w, http.StatusBadRequest, "line is required")
		return
	}
	if req.RiskLevel == "" {
		Error(w, http.StatusBadRequest, "riskLevel is required")
		return
	}

	ctx := r.Context()

	// Translate line string to ID
	lineID, err := h.metadataRepo.GetLineIDByName(ctx, req.Line)
	if err != nil {
		h.logger.Debug("create: invalid line", "line", req.Line, "error", err)
		Error(w, http.StatusBadRequest, "invalid line")
		return
	}

	// Translate riskLevel string to ID
	riskLevelID, err := h.metadataRepo.GetRiskLevelIDByName(ctx, req.RiskLevel)
	if err != nil {
		h.logger.Debug("create: invalid riskLevel", "riskLevel", req.RiskLevel, "error", err)
		Error(w, http.StatusBadRequest, "invalid riskLevel")
		return
	}

	// Translate optional category string to ID
	var categoryID *int
	if req.Category != nil && *req.Category != "" {
		id, err := h.metadataRepo.GetCategoryIDByName(ctx, *req.Category)
		if err != nil {
			h.logger.Debug("create: invalid category", "category", *req.Category, "error", err)
			Error(w, http.StatusBadRequest, "invalid category")
			return
		}
		categoryID = &id
	}

	// Translate optional responsible string to ID (by name, not username)
	var responsibleID *int16
	if req.Responsible != nil && *req.Responsible != "" {
		id, err := h.metadataRepo.GetResponsibleIDByName(ctx, *req.Responsible)
		if err != nil {
			h.logger.Debug("create: invalid responsible", "responsible", *req.Responsible, "error", err)
			Error(w, http.StatusBadRequest, "invalid responsible")
			return
		}
		responsibleID = &id
	}

	// Translate optional expenseType string to ID
	var expenseTypeID *int
	if req.ExpenseType != nil && *req.ExpenseType != "" {
		id, err := h.metadataRepo.GetExpenseTypeIDByName(ctx, *req.ExpenseType)
		if err != nil {
			h.logger.Debug("create: invalid expenseType", "expenseType", *req.ExpenseType, "error", err)
			Error(w, http.StatusBadRequest, "invalid expenseType")
			return
		}
		expenseTypeID = &id
	}

	// Parse MOC number as integer if provided
	var mocNumber *int
	if req.MOC != nil && *req.MOC != "" {
		num, err := strconv.Atoi(*req.MOC)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid moc format")
			return
		}
		mocNumber = &num
	}

	// Parse deadline if provided
	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		t, err := parseDate(*req.Deadline)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid deadline format, use YYYY-MM-DD or ISO 8601")
			return
		}
		deadline = &t
	}

	input := service.CreateRiskInput{
		RiskType:          req.RiskType,
		LinkedRiskID:      req.LinkedRiskID,
		Machines:          req.Machine,
		LineID:            lineID,
		CategoryID:        categoryID,
		DangerDescription: req.DescriptionOfDanger,
		RiskLevelID:       riskLevelID,
		Action:            req.Action,
		ResponsibleID:     responsibleID,
		MOCNumber:         mocNumber,
		Deadline:          deadline,
		Cost:              req.Cost,
		ExpenseTypeID:     expenseTypeID,
	}

	risk, err := h.riskService.Create(ctx, input, person.PersonID)
	if err != nil {
		switch err {
		case domain.ErrLinkedRiskSameType:
			Error(w, http.StatusBadRequest, "linked risk must be opposite type")
		case domain.ErrRiskNotFound:
			Error(w, http.StatusBadRequest, "linked risk not found")
		case domain.ErrValidationFailed:
			Error(w, http.StatusBadRequest, "validation failed")
		default:
			h.logger.Error("failed to create risk", "error", err)
			ErrorWithDetails(w, http.StatusInternalServerError, "failed to create risk", err.Error())
		}
		return
	}

	h.logger.Info("risk created", "id", risk.ID, "created_by", person.PersonID)
	JSON(w, http.StatusCreated, toRiskResponse(risk, true))
}

// Update handles PATCH /api/v1/entries/{id}
func (h *RiskHandler) Update(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("update: invalid request body", "error", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Debug("update: request received", "id", id, "machine", req.Machine, "line", req.Line,
		"category", req.Category, "riskLevel", req.RiskLevel, "status", req.Status,
		"responsible", req.Responsible, "moc", req.MOC, "expenseType", req.ExpenseType)

	ctx := r.Context()

	// Translate optional line string to ID
	var lineID *int16
	if req.Line != nil && *req.Line != "" {
		id, err := h.metadataRepo.GetLineIDByName(ctx, *req.Line)
		if err != nil {
			h.logger.Debug("update: invalid line", "line", *req.Line, "error", err)
			Error(w, http.StatusBadRequest, "invalid line")
			return
		}
		lineID = &id
	}

	// Translate optional category string to ID
	var categoryID *int
	if req.Category != nil && *req.Category != "" {
		id, err := h.metadataRepo.GetCategoryIDByName(ctx, *req.Category)
		if err != nil {
			h.logger.Debug("update: invalid category", "category", *req.Category, "error", err)
			Error(w, http.StatusBadRequest, "invalid category")
			return
		}
		categoryID = &id
	}

	// Translate optional riskLevel string to ID
	var riskLevelID *int16
	if req.RiskLevel != nil && *req.RiskLevel != "" {
		id, err := h.metadataRepo.GetRiskLevelIDByName(ctx, *req.RiskLevel)
		if err != nil {
			h.logger.Debug("update: invalid riskLevel", "riskLevel", *req.RiskLevel, "error", err)
			Error(w, http.StatusBadRequest, "invalid riskLevel")
			return
		}
		riskLevelID = &id
	}

	// Translate optional responsible string to ID (by name)
	var responsibleID *int16
	if req.Responsible != nil && *req.Responsible != "" {
		id, err := h.metadataRepo.GetResponsibleIDByName(ctx, *req.Responsible)
		if err != nil {
			h.logger.Debug("update: invalid responsible", "responsible", *req.Responsible, "error", err)
			Error(w, http.StatusBadRequest, "invalid responsible")
			return
		}
		responsibleID = &id
	}

	// Translate optional expenseType string to ID
	var expenseTypeID *int
	if req.ExpenseType != nil && *req.ExpenseType != "" {
		id, err := h.metadataRepo.GetExpenseTypeIDByName(ctx, *req.ExpenseType)
		if err != nil {
			h.logger.Debug("update: invalid expenseType", "expenseType", *req.ExpenseType, "error", err)
			Error(w, http.StatusBadRequest, "invalid expenseType")
			return
		}
		expenseTypeID = &id
	}

	// Translate optional status string to ID
	var actionStateID *int16
	if req.Status != nil && *req.Status != "" {
		id, err := h.metadataRepo.GetStatusIDByName(ctx, *req.Status)
		if err != nil {
			h.logger.Debug("update: invalid status", "status", *req.Status, "error", err)
			Error(w, http.StatusBadRequest, "invalid status")
			return
		}
		actionStateID = &id
	}

	// Parse MOC number as integer if provided
	var mocNumber *int
	if req.MOC != nil && *req.MOC != "" {
		num, err := strconv.Atoi(*req.MOC)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid moc format")
			return
		}
		mocNumber = &num
	}

	// Parse dates if provided
	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		t, err := parseDate(*req.Deadline)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid deadline format, use YYYY-MM-DD or ISO 8601")
			return
		}
		deadline = &t
	}

	var implDate *time.Time
	if req.ImplementationDate != nil && *req.ImplementationDate != "" {
		t, err := parseDate(*req.ImplementationDate)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid implementationDate format, use YYYY-MM-DD or ISO 8601")
			return
		}
		implDate = &t
	}

	input := service.UpdateRiskInput{
		Machines:             req.Machine,
		LineID:               lineID,
		CategoryID:           categoryID,
		DangerDescription:    req.DescriptionOfDanger,
		RiskLevelID:          riskLevelID,
		Action:               req.Action,
		ResponsibleID:        responsibleID,
		MOCNumber:            mocNumber,
		Deadline:             deadline,
		ActionStateID:        actionStateID,
		ImplementationDate:   implDate,
		Cost:                 req.Cost,
		ExpenseTypeID:        expenseTypeID,
		LinkedRiskID:         req.LinkedRiskID,
		StatusChangeReason:   req.StatusChangeReason,
		DeadlineChangeReason: req.DeadlineChangeReason,
	}

	risk, err := h.riskService.Update(ctx, id, input, person.PersonID)
	if err != nil {
		switch err {
		case domain.ErrRiskNotFound:
			Error(w, http.StatusNotFound, "risk not found")
		case domain.ErrLinkedRiskSameType:
			Error(w, http.StatusBadRequest, "linked risk must be opposite type")
		case domain.ErrRiskAlreadyClosed:
			Error(w, http.StatusBadRequest, "risk is already closed")
		case domain.ErrValidationFailed:
			Error(w, http.StatusBadRequest, "deadline change requires reason")
		default:
			h.logger.Error("failed to update risk", "id", id, "error", err)
			ErrorWithDetails(w, http.StatusInternalServerError, "failed to update risk", err.Error())
		}
		return
	}

	h.logger.Info("risk updated", "id", risk.ID, "updated_by", person.PersonID)
	JSON(w, http.StatusOK, toRiskResponse(risk, true))
}

// Delete handles DELETE /api/v1/entries/{id}
func (h *RiskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.riskService.Delete(r.Context(), id, person.PersonID); err != nil {
		if err == domain.ErrRiskNotFound {
			Error(w, http.StatusNotFound, "risk not found")
			return
		}
		h.logger.Error("failed to delete risk", "id", id, "error", err)
		ErrorWithDetails(w, http.StatusInternalServerError, "failed to delete risk", err.Error())
		return
	}

	h.logger.Info("risk deleted", "id", id, "deleted_by", person.PersonID)
	w.WriteHeader(http.StatusNoContent)
}

func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// parseDate tries multiple date formats: YYYY-MM-DD and RFC3339 (ISO 8601 with time)
func parseDate(s string) (time.Time, error) {
	// Try YYYY-MM-DD first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try RFC3339 (e.g., 2026-03-31T00:00:00Z)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

// statusChangeToAction converts a status name to an action verb.
func statusChangeToAction(newStatus string) string {
	switch newStatus {
	case "ASSIGNED":
		return "assigned"
	case "IN_PROGRESS":
		return "started"
	case "CLOSED":
		return "closed"
	case "OPEN":
		return "reopened"
	default:
		return "updated"
	}
}
