package repo

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

// PersonRepo handles person database operations.
type PersonRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPersonRepo creates a new person repository.
func NewPersonRepo(pool *pgxpool.Pool, logger *slog.Logger) *PersonRepo {
	return &PersonRepo{
		pool:   pool,
		logger: logger.With("component", "repo.person"),
	}
}

// GetByUsername fetches a person by username with role and permissions.
func (r *PersonRepo) GetByUsername(ctx context.Context, username string) (*domain.Person, error) {
	r.logger.Debug("fetching person by username", "username", username)

	var person domain.Person
	err := r.pool.QueryRow(ctx, personQ.GetByUsername, username).Scan(
		&person.PersonID,
		&person.PersonName,
		&person.PersonEmail,
		&person.Username,
		&person.PwHash,
		&person.RoleID,
		&person.Locked,
		&person.RoleName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Debug("person not found", "username", username)
			return nil, domain.ErrPersonNotFound
		}
		r.logger.Error("failed to query person", "username", username, "error", err)
		return nil, err
	}

	// Query permissions for this person's role
	if err := r.loadPermissions(ctx, &person); err != nil {
		return nil, err
	}

	r.logger.Debug("person loaded",
		"person_id", person.PersonID,
		"role", person.RoleName,
		"permissions_count", len(person.Permissions),
	)

	return &person, nil
}

// GetByID fetches a person by ID with role and permissions.
func (r *PersonRepo) GetByID(ctx context.Context, id int16) (*domain.Person, error) {
	r.logger.Debug("fetching person by id", "person_id", id)

	var person domain.Person
	err := r.pool.QueryRow(ctx, personQ.GetByID, id).Scan(
		&person.PersonID,
		&person.PersonName,
		&person.PersonEmail,
		&person.Username,
		&person.PwHash,
		&person.RoleID,
		&person.Locked,
		&person.RoleName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Debug("person not found", "person_id", id)
			return nil, domain.ErrPersonNotFound
		}
		r.logger.Error("failed to query person", "person_id", id, "error", err)
		return nil, err
	}

	// Query permissions for this person's role
	if err := r.loadPermissions(ctx, &person); err != nil {
		return nil, err
	}

	r.logger.Debug("person loaded",
		"person_id", person.PersonID,
		"role", person.RoleName,
		"permissions_count", len(person.Permissions),
	)

	return &person, nil
}

// CreateInitialAdmin inserts an admin user only if no persons exist yet.
// Returns false when users already exist (nothing is changed).
func (r *PersonRepo) CreateInitialAdmin(ctx context.Context, username, pwHash string) (bool, error) {
	var id int16
	err := r.pool.QueryRow(ctx, personQ.CreateInitialAdmin, username, pwHash).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Debug("users already exist, skipping initial admin")
			return false, nil
		}
		r.logger.Error("failed to create initial admin", "username", username, "error", err)
		return false, err
	}

	r.logger.Info("initial admin created", "person_id", id, "username", username)
	return true, nil
}

// loadPermissions loads permissions for a person's role.
func (r *PersonRepo) loadPermissions(ctx context.Context, person *domain.Person) error {
	rows, err := r.pool.Query(ctx, personQ.GetPermissions, person.RoleID)
	if err != nil {
		r.logger.Error("failed to query permissions", "role_id", person.RoleID, "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var perm domain.Permission
		if err := rows.Scan(&perm.PermissionID, &perm.Resource, &perm.Action); err != nil {
			return err
		}
		person.Permissions = append(person.Permissions, perm)
	}

	return rows.Err()
}
