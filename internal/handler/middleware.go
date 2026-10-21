package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/repo"
)

// ContextKey is a custom type for context keys.
type ContextKey string

const (
	// ContextKeyPerson is the key for storing Person in request context.
	ContextKeyPerson ContextKey = "person"
)

// AuthMiddleware provides authentication middleware.
type AuthMiddleware struct {
	sessionMgr *SessionManager
	personRepo repo.PersonRepository
	logger     *slog.Logger
}

// NewAuthMiddleware creates new auth middleware.
func NewAuthMiddleware(sessionMgr *SessionManager, personRepo repo.PersonRepository, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionMgr: sessionMgr,
		personRepo: personRepo,
		logger:     logger.With("component", "middleware.auth"),
	}
}

// RequireAuth ensures the request has a valid session and loads the Person into context.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.sessionMgr.Get(r)
		if err != nil {
			m.logger.Debug("auth rejected: session error", "error", err)
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		personID, ok := m.sessionMgr.GetPersonID(session)
		if !ok || session.IsNew {
			m.logger.Debug("auth rejected: no valid session")
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Load person from database to get current permissions
		person, err := m.personRepo.GetByID(r.Context(), personID)
		if err != nil {
			if err == domain.ErrPersonNotFound {
				m.logger.Debug("auth rejected: person not found", "person_id", personID)
				Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			m.logger.Error("auth error: failed to load person", "person_id", personID, "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if person.Locked {
			m.logger.Debug("auth rejected: account locked", "person_id", personID)
			Error(w, http.StatusForbidden, "account locked")
			return
		}

		m.logger.Debug("auth successful",
			"person_id", person.PersonID,
			"username", person.Username,
			"role", person.RoleName,
		)

		// Clear password hash
		person.PwHash = nil

		// Add person to context
		ctx := context.WithValue(r.Context(), ContextKeyPerson, person)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission checks if the authenticated user has the required permission.
func (m *AuthMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			person := GetPersonFromContext(r.Context())
			if person == nil {
				m.logger.Debug("permission check: no person in context")
				Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if !person.HasPermission(resource, action) {
				m.logger.Debug("permission denied",
					"person_id", person.PersonID,
					"resource", resource,
					"action", action,
				)
				Error(w, http.StatusForbidden, "forbidden")
				return
			}

			m.logger.Debug("permission granted",
				"person_id", person.PersonID,
				"resource", resource,
				"action", action,
			)

			next.ServeHTTP(w, r)
		})
	}
}

// GetPersonFromContext retrieves the Person from the request context.
func GetPersonFromContext(ctx context.Context) *domain.Person {
	val := ctx.Value(ContextKeyPerson)
	if val == nil {
		return nil
	}
	person, ok := val.(*domain.Person)
	if !ok {
		return nil
	}
	return person
}
