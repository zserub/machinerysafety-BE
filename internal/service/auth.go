package service

import (
	"context"
	"log/slog"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
	"github.com/zserub/machinerysafety-BE/internal/utils"
)

// AuthService handles authentication logic.
type AuthService struct {
	personRepo repo.PersonRepository
	logger     *slog.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(personRepo repo.PersonRepository, logger *slog.Logger) *AuthService {
	return &AuthService{
		personRepo: personRepo,
		logger:     logger.With("component", "service.auth"),
	}
}

// Login authenticates a user and returns their Person on success.
func (s *AuthService) Login(ctx context.Context, username, password string) (*domain.Person, error) {
	s.logger.Debug("login attempt", "username", username)

	// Fetch person from repo
	person, err := s.personRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == domain.ErrPersonNotFound {
			s.logger.Debug("login failed: user not found", "username", username)
			return nil, domain.ErrInvalidCredentials
		}
		s.logger.Error("login failed: repo error", "username", username, "error", err)
		return nil, err
	}

	// Check if account is locked
	if person.Locked {
		s.logger.Debug("login failed: account locked", "username", username, "person_id", person.PersonID)
		return nil, domain.ErrAccountLocked
	}

	// Verify password
	if person.PwHash == nil || !utils.VerifyPassword(password, *person.PwHash) {
		s.logger.Debug("login failed: invalid password", "username", username)
		return nil, domain.ErrInvalidCredentials
	}

	s.logger.Info("user logged in",
		"person_id", person.PersonID,
		"username", username,
		"role", person.RoleName,
	)

	// Clear password hash before returning
	person.PwHash = nil

	return person, nil
}
