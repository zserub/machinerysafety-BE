package repo

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// PersonRepository defines the interface for person data access.
type PersonRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.Person, error)
	GetByID(ctx context.Context, id int16) (*domain.Person, error)
}

// MetadataRepository defines the interface for metadata data access.
type MetadataRepository interface {
	GetAll(ctx context.Context) (*domain.Metadata, error)
}

// RiskRepository defines the interface for risk data access.
type RiskRepository interface {
	List(ctx context.Context, filters RiskFilters) (*RiskListResult, error)
	GetByID(ctx context.Context, id int64) (*domain.Risk, error)
	GetRiskType(ctx context.Context, id int64) (int, error)
	Create(ctx context.Context, tx pgx.Tx, risk *domain.Risk) error
	Update(ctx context.Context, tx pgx.Tx, risk *domain.Risk) error
	Delete(ctx context.Context, tx pgx.Tx, id int64) error
	InsertStatusHistory(ctx context.Context, tx pgx.Tx, item *domain.StatusHistoryItem) error
	InsertDeadlineHistory(ctx context.Context, tx pgx.Tx, item *domain.DeadlineHistoryItem) error
	SetAuditContext(ctx context.Context, tx pgx.Tx, userID int16) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

// PreferenceRepository defines the interface for preference data access.
type PreferenceRepository interface {
	GetByPersonID(ctx context.Context, personID int16) (*domain.Preference, error)
	Update(ctx context.Context, personID int16, updates *domain.PreferenceUpdate) (*domain.Preference, error)
}
