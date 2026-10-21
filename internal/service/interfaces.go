package service

import (
	"context"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// AuthServiceInterface defines the interface for authentication.
type AuthServiceInterface interface {
	Login(ctx context.Context, username, password string) (*domain.Person, error)
}

// RiskServiceInterface defines the interface for risk business logic.
type RiskServiceInterface interface {
	List(ctx context.Context, filters repo.RiskFilters) (*repo.RiskListResult, error)
	GetByID(ctx context.Context, id int64) (*domain.Risk, error)
	Create(ctx context.Context, input CreateRiskInput, createdByID int16) (*domain.Risk, error)
	Update(ctx context.Context, id int64, input UpdateRiskInput, updatedByID int16) (*domain.Risk, error)
	Delete(ctx context.Context, id int64, deletedByID int16) error
}
