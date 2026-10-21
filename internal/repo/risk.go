package repo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// RiskRepo handles risk database operations.
type RiskRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewRiskRepo creates a new risk repository.
func NewRiskRepo(pool *pgxpool.Pool, logger *slog.Logger) *RiskRepo {
	return &RiskRepo{
		pool:   pool,
		logger: logger.With("component", "repo.risk"),
	}
}

// RiskFilters defines filtering options for List.
type RiskFilters struct {
	Search        string
	RiskType      *int
	LineID        *int16
	CategoryID    *int
	RiskLevelID   *int16
	StatusID      *int16
	ResponsibleID *int16

	// Pagination
	Page     int
	PageSize int

	// Sorting
	SortBy    string // "creation_date", "deadline", "risk_level", etc.
	SortOrder string // "asc" or "desc"
}

// RiskListResult contains paginated results.
type RiskListResult struct {
	Risks      []domain.Risk `json:"entries"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}

// List fetches risks with filtering, sorting, and pagination.
// PageSize=0 means return all entries (for client-side pagination).
func (r *RiskRepo) List(ctx context.Context, filters RiskFilters) (*RiskListResult, error) {
	r.logger.Debug("listing risks", "filters", filters)

	// Default page to 1
	if filters.Page < 1 {
		filters.Page = 1
	}
	// PageSize=0 means "all entries" - don't enforce limits

	// Build WHERE clause
	var conditions []string
	var args []any
	argIdx := 1

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(r.machines ILIKE $%d OR r.danger_description ILIKE $%d OR r.action ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}
	if filters.RiskType != nil {
		conditions = append(conditions, fmt.Sprintf("r.risk_type = $%d", argIdx))
		args = append(args, *filters.RiskType)
		argIdx++
	}
	if filters.LineID != nil {
		conditions = append(conditions, fmt.Sprintf("r.line_id = $%d", argIdx))
		args = append(args, *filters.LineID)
		argIdx++
	}
	if filters.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("r.category_id = $%d", argIdx))
		args = append(args, *filters.CategoryID)
		argIdx++
	}
	if filters.RiskLevelID != nil {
		conditions = append(conditions, fmt.Sprintf("r.risk_level_id = $%d", argIdx))
		args = append(args, *filters.RiskLevelID)
		argIdx++
	}
	if filters.StatusID != nil {
		conditions = append(conditions, fmt.Sprintf("r.action_state_id = $%d", argIdx))
		args = append(args, *filters.StatusID)
		argIdx++
	}
	if filters.ResponsibleID != nil {
		conditions = append(conditions, fmt.Sprintf("r.responsible_id = $%d", argIdx))
		args = append(args, *filters.ResponsibleID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("%s %s", riskQ.CountRisks, whereClause)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		r.logger.Error("failed to count risks", "error", err)
		return nil, err
	}

	// Build ORDER BY clause
	orderBy := "r.creation_date DESC"
	if filters.SortBy != "" {
		sortCol := r.mapSortColumn(filters.SortBy)
		sortOrder := "ASC"
		if strings.ToLower(filters.SortOrder) == "desc" {
			sortOrder = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", sortCol, sortOrder)
	}

	// Main query with joins
	var query string
	if filters.PageSize > 0 {
		// Paginated query
		offset := (filters.Page - 1) * filters.PageSize
		query = fmt.Sprintf("%s %s ORDER BY %s LIMIT $%d OFFSET $%d",
			riskQ.ListRisksBase, whereClause, orderBy, argIdx, argIdx+1)
		args = append(args, filters.PageSize, offset)
	} else {
		// Return all entries (pageSize=0)
		query = fmt.Sprintf("%s %s ORDER BY %s", riskQ.ListRisksBase, whereClause, orderBy)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to query risks", "error", err)
		return nil, err
	}
	defer rows.Close()

	var risks []domain.Risk
	for rows.Next() {
		var risk domain.Risk
		if err := rows.Scan(
			&risk.ID, &risk.RiskType, &risk.LinkedRiskID, &risk.CreationDate, &risk.Machines,
			&risk.LineID, &risk.CategoryID, &risk.DangerDescription, &risk.RiskLevelID,
			&risk.Action, &risk.ResponsibleID, &risk.MOCNumber, &risk.Deadline, &risk.ActionStateID,
			&risk.ImplementationDate, &risk.EntryModifiedAt, &risk.CreatedByID, &risk.CommittedAt,
			&risk.Cost, &risk.ExpenseTypeID,
			&risk.LineName, &risk.CategoryName, &risk.RiskLevelName, &risk.StateName,
			&risk.ResponsibleName, &risk.CreatedByName, &risk.ExpenseTypeName,
		); err != nil {
			return nil, err
		}
		risks = append(risks, risk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination info for response
	pageSize := filters.PageSize
	totalPages := 1
	if pageSize > 0 {
		totalPages = int(total) / pageSize
		if int(total)%pageSize > 0 {
			totalPages++
		}
	} else {
		// When returning all entries, set pageSize to total for response
		pageSize = int(total)
	}

	r.logger.Debug("risks listed", "count", len(risks), "total", total)

	return &RiskListResult{
		Risks:      risks,
		Total:      total,
		Page:       filters.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *RiskRepo) mapSortColumn(sortBy string) string {
	switch sortBy {
	case "creation_date", "createdAt":
		return "r.creation_date"
	case "deadline":
		return "r.deadline"
	case "risk_level", "riskLevel":
		return "rl.severity"
	case "status":
		return "r.action_state_id"
	case "line":
		return "l.line_name"
	case "category":
		return "c.category_name"
	default:
		return "r.creation_date"
	}
}

// GetByID fetches a single risk with all related data.
func (r *RiskRepo) GetByID(ctx context.Context, id int64) (*domain.Risk, error) {
	r.logger.Debug("fetching risk by id", "id", id)

	var risk domain.Risk
	err := r.pool.QueryRow(ctx, riskQ.GetByID, id).Scan(
		&risk.ID, &risk.RiskType, &risk.LinkedRiskID, &risk.CreationDate, &risk.Machines,
		&risk.LineID, &risk.CategoryID, &risk.DangerDescription, &risk.RiskLevelID,
		&risk.Action, &risk.ResponsibleID, &risk.MOCNumber, &risk.Deadline, &risk.ActionStateID,
		&risk.ImplementationDate, &risk.EntryModifiedAt, &risk.CreatedByID, &risk.CommittedAt,
		&risk.Cost, &risk.ExpenseTypeID,
		&risk.LineName, &risk.CategoryName, &risk.RiskLevelName, &risk.StateName,
		&risk.ResponsibleName, &risk.CreatedByName, &risk.ExpenseTypeName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRiskNotFound
		}
		r.logger.Error("failed to query risk", "id", id, "error", err)
		return nil, err
	}

	// Load photos
	photos, err := r.getPhotos(ctx, id)
	if err != nil {
		return nil, err
	}
	risk.Photos = photos

	// Load status history
	statusHistory, err := r.getStatusHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	risk.StatusHistory = statusHistory

	// Load deadline history
	deadlineHistory, err := r.getDeadlineHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	risk.DeadlineHistory = deadlineHistory

	// Load linked risk summary if exists
	if risk.LinkedRiskID != nil {
		linkedRisk, err := r.getLinkedRiskSummary(ctx, *risk.LinkedRiskID)
		if err != nil && err != domain.ErrRiskNotFound {
			return nil, err
		}
		risk.LinkedRisk = linkedRisk
	}

	r.logger.Debug("risk loaded", "id", id)
	return &risk, nil
}

func (r *RiskRepo) getPhotos(ctx context.Context, riskID int64) ([]domain.Photo, error) {
	rows, err := r.pool.Query(ctx, riskQ.GetPhotos, riskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []domain.Photo
	for rows.Next() {
		var p domain.Photo
		if err := rows.Scan(&p.RiskID, &p.PhotoNumber, &p.FilePath, &p.UploadedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}

	return photos, rows.Err()
}

func (r *RiskRepo) getStatusHistory(ctx context.Context, riskID int64) ([]domain.StatusHistoryItem, error) {
	rows, err := r.pool.Query(ctx, riskQ.GetStatusHistory, riskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.StatusHistoryItem
	for rows.Next() {
		var h domain.StatusHistoryItem
		if err := rows.Scan(
			&h.ID, &h.RiskID, &h.OldStatus, &h.NewStatus, &h.Reason,
			&h.ChangedBy, &h.ChangedAt,
			&h.OldStateName, &h.NewStateName, &h.ChangedByName,
		); err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

func (r *RiskRepo) getDeadlineHistory(ctx context.Context, riskID int64) ([]domain.DeadlineHistoryItem, error) {
	rows, err := r.pool.Query(ctx, riskQ.GetDeadlineHistory, riskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.DeadlineHistoryItem
	for rows.Next() {
		var h domain.DeadlineHistoryItem
		if err := rows.Scan(
			&h.ID, &h.RiskID, &h.OldDeadline, &h.NewDeadline, &h.Reason,
			&h.ChangedBy, &h.ChangedAt, &h.ChangedByName,
		); err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

func (r *RiskRepo) getLinkedRiskSummary(ctx context.Context, id int64) (*domain.LinkedRiskSummary, error) {
	var lr domain.LinkedRiskSummary
	err := r.pool.QueryRow(ctx, riskQ.GetLinkedRiskSummary, id).Scan(
		&lr.ID, &lr.Machines, &lr.LineName, &lr.DangerDescription, &lr.RiskType,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRiskNotFound
		}
		return nil, err
	}

	return &lr, nil
}

// Create inserts a new risk.
func (r *RiskRepo) Create(ctx context.Context, tx pgx.Tx, risk *domain.Risk) error {
	err := tx.QueryRow(ctx, riskQ.Create,
		risk.RiskType, risk.LinkedRiskID, risk.Machines, risk.LineID, risk.CategoryID,
		risk.DangerDescription, risk.RiskLevelID, risk.Action, risk.ResponsibleID, risk.MOCNumber,
		risk.Deadline, risk.ActionStateID, risk.CreatedByID, risk.Cost, risk.ExpenseTypeID,
	).Scan(&risk.ID, &risk.CreationDate, &risk.EntryModifiedAt)

	if err != nil {
		r.logger.Error("failed to create risk", "error", err)
		return err
	}

	r.logger.Info("risk created", "id", risk.ID, "created_by", risk.CreatedByID)
	return nil
}

// Update modifies an existing risk.
func (r *RiskRepo) Update(ctx context.Context, tx pgx.Tx, risk *domain.Risk) error {
	err := tx.QueryRow(ctx, riskQ.Update,
		risk.ID, risk.Machines, risk.LineID, risk.CategoryID, risk.DangerDescription,
		risk.RiskLevelID, risk.Action, risk.ResponsibleID, risk.MOCNumber,
		risk.Deadline, risk.ActionStateID, risk.ImplementationDate,
		risk.Cost, risk.ExpenseTypeID, risk.LinkedRiskID,
	).Scan(&risk.EntryModifiedAt)

	if err != nil {
		r.logger.Error("failed to update risk", "id", risk.ID, "error", err)
		return err
	}

	r.logger.Info("risk updated", "id", risk.ID)
	return nil
}

// Delete removes a risk and all related records.
func (r *RiskRepo) Delete(ctx context.Context, tx pgx.Tx, id int64) error {
	// Delete related records first (foreign key dependencies)
	if _, err := tx.Exec(ctx, riskQ.DeleteStatusHistory, id); err != nil {
		r.logger.Error("failed to delete status history", "id", id, "error", err)
		return err
	}
	if _, err := tx.Exec(ctx, riskQ.DeleteDeadlineHistory, id); err != nil {
		r.logger.Error("failed to delete deadline history", "id", id, "error", err)
		return err
	}
	if _, err := tx.Exec(ctx, riskQ.DeleteComments, id); err != nil {
		r.logger.Error("failed to delete comments", "id", id, "error", err)
		return err
	}
	if _, err := tx.Exec(ctx, riskQ.DeletePhotos, id); err != nil {
		r.logger.Error("failed to delete photos", "id", id, "error", err)
		return err
	}
	// Clear linked_risk_id references from other risks
	if _, err := tx.Exec(ctx, riskQ.ClearLinkedRiskRefs, id); err != nil {
		r.logger.Error("failed to clear linked risk references", "id", id, "error", err)
		return err
	}

	// Delete the risk itself
	result, err := tx.Exec(ctx, riskQ.Delete, id)
	if err != nil {
		r.logger.Error("failed to delete risk", "id", id, "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRiskNotFound
	}

	r.logger.Info("risk deleted", "id", id)
	return nil
}

// GetRiskType fetches only the risk_type of a risk (for validation).
func (r *RiskRepo) GetRiskType(ctx context.Context, id int64) (int, error) {
	var riskType int
	err := r.pool.QueryRow(ctx, riskQ.GetRiskType, id).Scan(&riskType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, domain.ErrRiskNotFound
		}
		return 0, err
	}

	return riskType, nil
}

// GetCurrentStatus fetches the current status of a risk.
func (r *RiskRepo) GetCurrentStatus(ctx context.Context, id int64) (int16, error) {
	var status int16
	err := r.pool.QueryRow(ctx, riskQ.GetCurrentStatus, id).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, domain.ErrRiskNotFound
		}
		return 0, err
	}

	return status, nil
}

// InsertStatusHistory records a status change.
func (r *RiskRepo) InsertStatusHistory(ctx context.Context, tx pgx.Tx, item *domain.StatusHistoryItem) error {
	err := tx.QueryRow(ctx, riskQ.InsertStatusHistory,
		item.RiskID, item.OldStatus, item.NewStatus, item.Reason, item.ChangedBy,
	).Scan(&item.ID, &item.ChangedAt)

	if err != nil {
		r.logger.Error("failed to insert status history", "risk_id", item.RiskID, "error", err)
		return err
	}

	r.logger.Debug("status history recorded", "risk_id", item.RiskID, "old", item.OldStatus, "new", item.NewStatus)
	return nil
}

// InsertDeadlineHistory records a deadline change.
func (r *RiskRepo) InsertDeadlineHistory(ctx context.Context, tx pgx.Tx, item *domain.DeadlineHistoryItem) error {
	err := tx.QueryRow(ctx, riskQ.InsertDeadlineHistory,
		item.RiskID, item.OldDeadline, item.NewDeadline, item.Reason, item.ChangedBy,
	).Scan(&item.ID, &item.ChangedAt)

	if err != nil {
		r.logger.Error("failed to insert deadline history", "risk_id", item.RiskID, "error", err)
		return err
	}

	r.logger.Debug("deadline history recorded", "risk_id", item.RiskID)
	return nil
}

// SetAuditContext sets the app.current_user_id for audit triggers.
func (r *RiskRepo) SetAuditContext(ctx context.Context, tx pgx.Tx, userID int16) error {
	// SET LOCAL doesn't support parameterized values, so we format directly.
	// userID is int16, so no SQL injection risk.
	_, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%d'", userID))
	return err
}

// BeginTx starts a new transaction.
func (r *RiskRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// GetCurrentDeadline fetches the current deadline of a risk.
func (r *RiskRepo) GetCurrentDeadline(ctx context.Context, id int64) (*time.Time, error) {
	var deadline *time.Time
	err := r.pool.QueryRow(ctx, riskQ.GetCurrentDeadline, id).Scan(&deadline)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRiskNotFound
		}
		return nil, err
	}

	return deadline, nil
}
