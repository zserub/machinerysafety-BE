package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// RiskService handles risk business logic.
type RiskService struct {
	riskRepo *repo.RiskRepo
	logger   *slog.Logger
}

// NewRiskService creates a new risk service.
func NewRiskService(riskRepo *repo.RiskRepo, logger *slog.Logger) *RiskService {
	return &RiskService{
		riskRepo: riskRepo,
		logger:   logger.With("component", "service.risk"),
	}
}

// CreateRiskInput represents input for creating a risk.
type CreateRiskInput struct {
	RiskType          int
	LinkedRiskID      *int64
	Machines          string
	LineID            int16
	CategoryID        *int
	DangerDescription string
	RiskLevelID       int16
	Action            *string
	ResponsibleID     *int16
	MOCNumber         *int
	Deadline          *time.Time
	Cost              *float64
	ExpenseTypeID     *int
}

// UpdateRiskInput represents input for updating a risk.
type UpdateRiskInput struct {
	Machines           *string
	LineID             *int16
	CategoryID         *int
	DangerDescription  *string
	RiskLevelID        *int16
	Action             *string
	ResponsibleID      *int16
	MOCNumber          *int
	Deadline           *time.Time
	ActionStateID      *int16
	ImplementationDate *time.Time
	Cost               *float64
	ExpenseTypeID      *int
	LinkedRiskID       *int64

	// Reason fields for tracked changes
	StatusChangeReason   *string
	DeadlineChangeReason *string
}

// List lists risks with filters.
func (s *RiskService) List(ctx context.Context, filters repo.RiskFilters) (*repo.RiskListResult, error) {
	return s.riskRepo.List(ctx, filters)
}

// GetByID gets a risk by ID.
func (s *RiskService) GetByID(ctx context.Context, id int64) (*domain.Risk, error) {
	return s.riskRepo.GetByID(ctx, id)
}

// Create creates a new risk with validation.
func (s *RiskService) Create(ctx context.Context, input CreateRiskInput, createdByID int16) (*domain.Risk, error) {
	s.logger.Debug("creating risk", "created_by", createdByID)

	// Validate linked risk type rule
	if input.LinkedRiskID != nil {
		linkedType, err := s.riskRepo.GetRiskType(ctx, *input.LinkedRiskID)
		if err != nil {
			if err == domain.ErrRiskNotFound {
				s.logger.Debug("linked risk not found", "linked_id", *input.LinkedRiskID)
				return nil, domain.ErrRiskNotFound
			}
			return nil, err
		}
		if linkedType == input.RiskType {
			s.logger.Debug("linked risk same type", "type", input.RiskType)
			return nil, domain.ErrLinkedRiskSameType
		}
	}

	// Begin transaction
	tx, err := s.riskRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Set audit context
	if err := s.riskRepo.SetAuditContext(ctx, tx, createdByID); err != nil {
		return nil, err
	}

	// Create risk
	risk := &domain.Risk{
		RiskType:          input.RiskType,
		LinkedRiskID:      input.LinkedRiskID,
		Machines:          input.Machines,
		LineID:            input.LineID,
		CategoryID:        input.CategoryID,
		DangerDescription: input.DangerDescription,
		RiskLevelID:       input.RiskLevelID,
		Action:            input.Action,
		ResponsibleID:     input.ResponsibleID,
		MOCNumber:         input.MOCNumber,
		Deadline:          input.Deadline,
		ActionStateID:     1, // OPEN
		CreatedByID:       createdByID,
		Cost:              input.Cost,
		ExpenseTypeID:     input.ExpenseTypeID,
	}

	if err := s.riskRepo.Create(ctx, tx, risk); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("risk created", "id", risk.ID, "created_by", createdByID)

	// Return full risk with joined data
	return s.riskRepo.GetByID(ctx, risk.ID)
}

// Update updates a risk with status/deadline history tracking.
func (s *RiskService) Update(ctx context.Context, id int64, input UpdateRiskInput, updatedByID int16) (*domain.Risk, error) {
	s.logger.Debug("updating risk", "id", id, "updated_by", updatedByID)

	// Fetch existing risk
	existing, err := s.riskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already closed
	if existing.ActionStateID == 7 { // CLOSED (state_id 7)
		s.logger.Debug("risk already closed", "id", id)
		return nil, domain.ErrRiskAlreadyClosed
	}

	// Validate linked risk type if changing
	if input.LinkedRiskID != nil && (existing.LinkedRiskID == nil || *input.LinkedRiskID != *existing.LinkedRiskID) {
		linkedType, err := s.riskRepo.GetRiskType(ctx, *input.LinkedRiskID)
		if err != nil {
			if err == domain.ErrRiskNotFound {
				return nil, domain.ErrRiskNotFound
			}
			return nil, err
		}
		if linkedType == existing.RiskType {
			s.logger.Debug("linked risk same type", "type", existing.RiskType)
			return nil, domain.ErrLinkedRiskSameType
		}
	}

	// Begin transaction
	tx, err := s.riskRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Set audit context
	if err := s.riskRepo.SetAuditContext(ctx, tx, updatedByID); err != nil {
		return nil, err
	}

	// Track status change
	s.logger.Debug("checking status change",
		"input_state_id", input.ActionStateID,
		"existing_state_id", existing.ActionStateID)
	if input.ActionStateID != nil && *input.ActionStateID != existing.ActionStateID {
		s.logger.Debug("inserting status history",
			"old_status", existing.ActionStateID,
			"new_status", *input.ActionStateID)
		statusHistory := &domain.StatusHistoryItem{
			RiskID:    id,
			OldStatus: existing.ActionStateID,
			NewStatus: *input.ActionStateID,
			ChangedBy: updatedByID,
			Reason:    input.StatusChangeReason,
		}
		if err := s.riskRepo.InsertStatusHistory(ctx, tx, statusHistory); err != nil {
			return nil, err
		}
	}

	// Track deadline change
	if input.Deadline != nil && existing.Deadline != nil && !input.Deadline.Equal(*existing.Deadline) {
		if input.DeadlineChangeReason == nil || *input.DeadlineChangeReason == "" {
			s.logger.Debug("deadline change requires reason", "id", id)
			return nil, domain.ErrValidationFailed
		}
		deadlineHistory := &domain.DeadlineHistoryItem{
			RiskID:      id,
			OldDeadline: *existing.Deadline,
			NewDeadline: *input.Deadline,
			Reason:      *input.DeadlineChangeReason,
			ChangedBy:   updatedByID,
		}
		if err := s.riskRepo.InsertDeadlineHistory(ctx, tx, deadlineHistory); err != nil {
			return nil, err
		}
	}

	// Apply updates to existing risk
	risk := s.applyUpdates(existing, input)

	// Update
	if err := s.riskRepo.Update(ctx, tx, risk); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("risk updated", "id", id, "updated_by", updatedByID)

	// Return full risk with joined data
	return s.riskRepo.GetByID(ctx, id)
}

func (s *RiskService) applyUpdates(existing *domain.Risk, input UpdateRiskInput) *domain.Risk {
	risk := *existing // Copy

	if input.Machines != nil {
		risk.Machines = *input.Machines
	}
	if input.LineID != nil {
		risk.LineID = *input.LineID
	}
	if input.CategoryID != nil {
		risk.CategoryID = input.CategoryID
	}
	if input.DangerDescription != nil {
		risk.DangerDescription = *input.DangerDescription
	}
	if input.RiskLevelID != nil {
		risk.RiskLevelID = *input.RiskLevelID
	}
	if input.Action != nil {
		risk.Action = input.Action
	}
	if input.ResponsibleID != nil {
		risk.ResponsibleID = input.ResponsibleID
	}
	if input.MOCNumber != nil {
		risk.MOCNumber = input.MOCNumber
	}
	if input.Deadline != nil {
		risk.Deadline = input.Deadline
	}
	if input.ActionStateID != nil {
		risk.ActionStateID = *input.ActionStateID
	}
	if input.ImplementationDate != nil {
		risk.ImplementationDate = input.ImplementationDate
	}
	if input.Cost != nil {
		risk.Cost = input.Cost
	}
	if input.ExpenseTypeID != nil {
		risk.ExpenseTypeID = input.ExpenseTypeID
	}
	if input.LinkedRiskID != nil {
		risk.LinkedRiskID = input.LinkedRiskID
	}

	return &risk
}

// Delete deletes a risk.
func (s *RiskService) Delete(ctx context.Context, id int64, deletedByID int16) error {
	s.logger.Debug("deleting risk", "id", id, "deleted_by", deletedByID)

	// Check if risk exists
	_, err := s.riskRepo.GetRiskType(ctx, id)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := s.riskRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Set audit context
	if err := s.riskRepo.SetAuditContext(ctx, tx, deletedByID); err != nil {
		return err
	}

	// Delete
	if err := s.riskRepo.Delete(ctx, tx, id); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.logger.Info("risk deleted", "id", id, "deleted_by", deletedByID)
	return nil
}
