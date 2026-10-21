package repo

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// PreferenceRepo handles preference database operations.
type PreferenceRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPreferenceRepo creates a new preference repository.
func NewPreferenceRepo(pool *pgxpool.Pool, logger *slog.Logger) *PreferenceRepo {
	return &PreferenceRepo{
		pool:   pool,
		logger: logger.With("component", "repo.preference"),
	}
}

// GetByPersonID fetches preferences for a person. Creates default if not exists.
func (r *PreferenceRepo) GetByPersonID(ctx context.Context, personID int16) (*domain.Preference, error) {
	r.logger.Debug("fetching preferences", "person_id", personID)

	var pref domain.Preference
	err := r.pool.QueryRow(ctx, preferenceQ.GetByPersonID, personID).Scan(
		&pref.PersonID,
		&pref.Language,
		&pref.Theme,
		&pref.ActiveColumns,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create default preferences
			r.logger.Debug("creating default preferences", "person_id", personID)
			return r.createDefault(ctx, personID)
		}
		r.logger.Error("failed to query preferences", "person_id", personID, "error", err)
		return nil, err
	}

	r.logger.Debug("preferences loaded", "person_id", personID, "language", pref.Language, "theme", pref.Theme)
	return &pref, nil
}

// Update updates preferences for a person.
func (r *PreferenceRepo) Update(ctx context.Context, personID int16, updates *domain.PreferenceUpdate) (*domain.Preference, error) {
	r.logger.Debug("updating preferences", "person_id", personID)

	// First ensure preferences exist
	_, err := r.GetByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}

	var pref domain.Preference

	// Use appropriate query based on whether columns are being updated
	if updates.Columns != nil {
		err = r.pool.QueryRow(ctx, preferenceQ.UpdateWithColumns, personID, updates.Language, updates.Theme, updates.Columns).Scan(
			&pref.PersonID,
			&pref.Language,
			&pref.Theme,
			&pref.ActiveColumns,
		)
	} else {
		err = r.pool.QueryRow(ctx, preferenceQ.Update, personID, updates.Language, updates.Theme).Scan(
			&pref.PersonID,
			&pref.Language,
			&pref.Theme,
			&pref.ActiveColumns,
		)
	}

	if err != nil {
		r.logger.Error("failed to update preferences", "person_id", personID, "error", err)
		return nil, err
	}

	r.logger.Info("preferences updated", "person_id", personID, "language", pref.Language, "theme", pref.Theme)
	return &pref, nil
}

// createDefault creates default preferences for a person.
func (r *PreferenceRepo) createDefault(ctx context.Context, personID int16) (*domain.Preference, error) {
	var pref domain.Preference
	err := r.pool.QueryRow(ctx, preferenceQ.CreateDefault, personID).Scan(
		&pref.PersonID,
		&pref.Language,
		&pref.Theme,
		&pref.ActiveColumns,
	)
	if err != nil {
		r.logger.Error("failed to create default preferences", "person_id", personID, "error", err)
		return nil, err
	}

	r.logger.Info("default preferences created", "person_id", personID)
	return &pref, nil
}
