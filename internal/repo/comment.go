package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// CommentRepo handles comment database operations.
type CommentRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewCommentRepo creates a new comment repository.
func NewCommentRepo(pool *pgxpool.Pool, logger *slog.Logger) *CommentRepo {
	return &CommentRepo{
		pool:   pool,
		logger: logger.With("component", "repo.comment"),
	}
}

// ListByRiskID fetches all comments for a given risk, ordered by creation time.
func (r *CommentRepo) ListByRiskID(ctx context.Context, riskID int64) ([]domain.Comment, error) {
	r.logger.Debug("listing comments", "risk_id", riskID)

	rows, err := r.pool.Query(ctx, commentQ.ListByRiskID, riskID)
	if err != nil {
		r.logger.Error("failed to query comments", "risk_id", riskID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		var authorName *string
		if err := rows.Scan(
			&c.ID,
			&c.RiskID,
			&c.AuthorID,
			&c.Content,
			&c.CreatedAt,
			&c.EditedAt,
			&authorName,
		); err != nil {
			r.logger.Error("failed to scan comment", "error", err)
			return nil, err
		}
		if authorName != nil {
			c.AuthorName = *authorName
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating comments", "error", err)
		return nil, err
	}

	r.logger.Debug("listed comments", "risk_id", riskID, "count", len(comments))
	return comments, nil
}

// Create inserts a new comment within the given transaction.
func (r *CommentRepo) Create(ctx context.Context, tx pgx.Tx, comment *domain.Comment) error {
	r.logger.Debug("creating comment", "risk_id", comment.RiskID, "author_id", comment.AuthorID)

	err := tx.QueryRow(ctx, commentQ.Create,
		comment.RiskID,
		comment.AuthorID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		r.logger.Error("failed to create comment", "error", err)
		return err
	}

	r.logger.Debug("created comment", "id", comment.ID)
	return nil
}

// GetByID fetches a comment by ID.
func (r *CommentRepo) GetByID(ctx context.Context, id int64) (*domain.Comment, error) {
	r.logger.Debug("fetching comment", "id", id)

	var c domain.Comment
	var authorName *string
	err := r.pool.QueryRow(ctx, commentQ.GetByID, id).Scan(
		&c.ID,
		&c.RiskID,
		&c.AuthorID,
		&c.Content,
		&c.CreatedAt,
		&c.EditedAt,
		&authorName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrCommentNotFound
		}
		r.logger.Error("failed to fetch comment", "id", id, "error", err)
		return nil, err
	}
	if authorName != nil {
		c.AuthorName = *authorName
	}

	return &c, nil
}

// BeginTx starts a new transaction.
func (r *CommentRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// SetAuditContext sets the current user ID for audit triggers.
func (r *CommentRepo) SetAuditContext(ctx context.Context, tx pgx.Tx, userID int16) error {
	// SET LOCAL doesn't support parameterized values, so we format directly.
	// userID is int16, so no SQL injection risk.
	_, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%d'", userID))
	return err
}
