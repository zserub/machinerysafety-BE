package repo

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// MetadataRepo handles metadata database operations.
type MetadataRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewMetadataRepo creates a new metadata repository.
func NewMetadataRepo(pool *pgxpool.Pool, logger *slog.Logger) *MetadataRepo {
	return &MetadataRepo{
		pool:   pool,
		logger: logger.With("component", "repo.metadata"),
	}
}

// GetAll fetches all reference data for dropdowns.
func (r *MetadataRepo) GetAll(ctx context.Context) (*domain.Metadata, error) {
	r.logger.Debug("fetching all metadata")

	metadata := &domain.Metadata{}

	// Fetch risk levels
	riskLevels, err := r.getRiskLevels(ctx)
	if err != nil {
		return nil, err
	}
	metadata.RiskLevels = riskLevels

	// Fetch statuses
	statuses, err := r.getStatuses(ctx)
	if err != nil {
		return nil, err
	}
	metadata.Statuses = statuses

	// Fetch categories
	categories, err := r.getCategories(ctx)
	if err != nil {
		return nil, err
	}
	metadata.Categories = categories

	// Fetch expense types
	expenseTypes, err := r.getExpenseTypes(ctx)
	if err != nil {
		return nil, err
	}
	metadata.ExpenseTypes = expenseTypes

	// Fetch lines
	lines, err := r.getLines(ctx)
	if err != nil {
		return nil, err
	}
	metadata.Lines = lines

	// Fetch responsibles
	responsibles, err := r.getResponsibles(ctx)
	if err != nil {
		return nil, err
	}
	metadata.Responsibles = responsibles

	r.logger.Debug("metadata loaded",
		"risk_levels", len(metadata.RiskLevels),
		"statuses", len(metadata.Statuses),
		"categories", len(metadata.Categories),
		"expense_types", len(metadata.ExpenseTypes),
		"lines", len(metadata.Lines),
		"responsibles", len(metadata.Responsibles),
	)

	return metadata, nil
}

func (r *MetadataRepo) getRiskLevels(ctx context.Context) ([]domain.RiskLevelOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetRiskLevels)
	if err != nil {
		r.logger.Error("failed to query risk levels", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.RiskLevelOption
	for rows.Next() {
		var opt domain.RiskLevelOption
		if err := rows.Scan(&opt.Value, &opt.Order); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

func (r *MetadataRepo) getStatuses(ctx context.Context) ([]domain.StatusOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetStatuses)
	if err != nil {
		r.logger.Error("failed to query statuses", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.StatusOption
	for rows.Next() {
		var opt domain.StatusOption
		if err := rows.Scan(&opt.Value, &opt.Order, &opt.IsTerminal); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

func (r *MetadataRepo) getCategories(ctx context.Context) ([]domain.CategoryOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetCategories)
	if err != nil {
		r.logger.Error("failed to query categories", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.CategoryOption
	for rows.Next() {
		var opt domain.CategoryOption
		if err := rows.Scan(&opt.Value, &opt.Order); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

func (r *MetadataRepo) getExpenseTypes(ctx context.Context) ([]domain.ExpenseTypeOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetExpenseTypes)
	if err != nil {
		r.logger.Error("failed to query expense types", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.ExpenseTypeOption
	for rows.Next() {
		var opt domain.ExpenseTypeOption
		if err := rows.Scan(&opt.Value, &opt.Order); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

func (r *MetadataRepo) getLines(ctx context.Context) ([]domain.LineOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetLines)
	if err != nil {
		r.logger.Error("failed to query lines", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.LineOption
	for rows.Next() {
		var opt domain.LineOption
		if err := rows.Scan(&opt.Value, &opt.Order); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

func (r *MetadataRepo) getResponsibles(ctx context.Context) ([]domain.ResponsibleOption, error) {
	rows, err := r.pool.Query(ctx, metadataQ.GetResponsibles)
	if err != nil {
		r.logger.Error("failed to query responsibles", "error", err)
		return nil, err
	}
	defer rows.Close()

	var result []domain.ResponsibleOption
	for rows.Next() {
		var opt domain.ResponsibleOption
		if err := rows.Scan(&opt.Value, &opt.Name, &opt.Order); err != nil {
			return nil, err
		}
		result = append(result, opt)
	}

	return result, rows.Err()
}

// Lookup functions to translate string values to IDs
// All lookups are case-insensitive for better frontend compatibility.

// GetLineIDByName returns the line_id for a given line_name.
func (r *MetadataRepo) GetLineIDByName(ctx context.Context, name string) (int16, error) {
	var id int16
	err := r.pool.QueryRow(ctx, metadataQ.GetLineIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("line not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}

// GetCategoryIDByName returns the id for a given category_name.
func (r *MetadataRepo) GetCategoryIDByName(ctx context.Context, name string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, metadataQ.GetCategoryIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("category not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}

// GetRiskLevelIDByName returns the risk_level_id for a given risk_level name.
func (r *MetadataRepo) GetRiskLevelIDByName(ctx context.Context, name string) (int16, error) {
	var id int16
	err := r.pool.QueryRow(ctx, metadataQ.GetRiskLevelIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("risk level not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}

// GetExpenseTypeIDByName returns the id for a given expense type name.
func (r *MetadataRepo) GetExpenseTypeIDByName(ctx context.Context, name string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, metadataQ.GetExpenseTypeIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("expense type not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}

// GetResponsibleIDByUsername returns the person_id for a given username.
func (r *MetadataRepo) GetResponsibleIDByUsername(ctx context.Context, username string) (int16, error) {
	var id int16
	err := r.pool.QueryRow(ctx, metadataQ.GetResponsibleIDByUsername, username).Scan(&id)
	if err != nil {
		r.logger.Debug("responsible not found", "username", username, "error", err)
		return 0, err
	}
	return id, nil
}

// GetResponsibleIDByName returns the person_id for a given person_name (display name).
func (r *MetadataRepo) GetResponsibleIDByName(ctx context.Context, name string) (int16, error) {
	var id int16
	err := r.pool.QueryRow(ctx, metadataQ.GetResponsibleIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("responsible not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}

// GetStatusIDByName returns the state_id for a given state name.
func (r *MetadataRepo) GetStatusIDByName(ctx context.Context, name string) (int16, error) {
	var id int16
	err := r.pool.QueryRow(ctx, metadataQ.GetStatusIDByName, name).Scan(&id)
	if err != nil {
		r.logger.Debug("status not found", "name", name, "error", err)
		return 0, err
	}
	return id, nil
}
