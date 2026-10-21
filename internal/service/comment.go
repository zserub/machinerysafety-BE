package service

import (
	"context"
	"log/slog"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// CommentService handles comment business logic.
type CommentService struct {
	commentRepo *repo.CommentRepo
	riskRepo    *repo.RiskRepo
	logger      *slog.Logger
}

// NewCommentService creates a new comment service.
func NewCommentService(commentRepo *repo.CommentRepo, riskRepo *repo.RiskRepo, logger *slog.Logger) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		riskRepo:    riskRepo,
		logger:      logger.With("component", "service.comment"),
	}
}

// CreateCommentInput represents input for creating a comment.
type CreateCommentInput struct {
	RiskID  int64
	Content string
}

// List returns all comments for a given risk.
func (s *CommentService) List(ctx context.Context, riskID int64) ([]domain.Comment, error) {
	s.logger.Debug("listing comments", "risk_id", riskID)

	// Verify risk exists
	_, err := s.riskRepo.GetRiskType(ctx, riskID)
	if err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.ListByRiskID(ctx, riskID)
	if err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for JSON serialization
	if comments == nil {
		comments = []domain.Comment{}
	}

	return comments, nil
}

// Create creates a new comment.
func (s *CommentService) Create(ctx context.Context, input CreateCommentInput, authorID int16) (*domain.Comment, error) {
	s.logger.Debug("creating comment", "risk_id", input.RiskID, "author_id", authorID)

	// Verify risk exists
	_, err := s.riskRepo.GetRiskType(ctx, input.RiskID)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	tx, err := s.commentRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Set audit context
	if err := s.commentRepo.SetAuditContext(ctx, tx, authorID); err != nil {
		return nil, err
	}

	// Create comment
	comment := &domain.Comment{
		RiskID:   input.RiskID,
		AuthorID: authorID,
		Content:  input.Content,
	}

	if err := s.commentRepo.Create(ctx, tx, comment); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("comment created", "id", comment.ID, "risk_id", input.RiskID, "author_id", authorID)

	// Fetch the full comment with author name
	return s.commentRepo.GetByID(ctx, comment.ID)
}
