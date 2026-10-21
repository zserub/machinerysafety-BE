package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/service"
)

// LoginRequest is the request body for login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the response for successful login.
type LoginResponse struct {
	PersonID    int16   `json:"person_id"`
	PersonName  *string `json:"person_name"`
	PersonEmail *string `json:"person_email"`
	Username    *string `json:"username"`
	RoleName    string  `json:"role_name"`
	Permissions []string `json:"permissions"`
}

// UserPermissionsResponse is the response for /user/permissions endpoint.
type UserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService service.AuthServiceInterface
	sessionMgr  *SessionManager
	logger      *slog.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService service.AuthServiceInterface, sessionMgr *SessionManager, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		sessionMgr:  sessionMgr,
		logger:      logger.With("component", "handler.auth"),
	}
}

// Login handles POST /api/v1/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("login: invalid request body", "error", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "username and password required")
		return
	}

	person, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			Error(w, http.StatusUnauthorized, "invalid credentials")
		case domain.ErrAccountLocked:
			Error(w, http.StatusForbidden, "account locked")
		default:
			h.logger.Error("login error", "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Create session
	session, err := h.sessionMgr.Get(r)
	if err != nil {
		h.logger.Error("failed to get session", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.sessionMgr.SetPersonID(session, person.PersonID)
	if person.Username != nil {
		h.sessionMgr.SetUsername(session, *person.Username)
	}

	if err := h.sessionMgr.Save(r, w, session); err != nil {
		h.logger.Error("failed to save session", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("user logged in",
		"person_id", person.PersonID,
		"username", person.Username,
	)

	JSON(w, http.StatusOK, LoginResponse{
		PersonID:    person.PersonID,
		PersonName:  person.PersonName,
		PersonEmail: person.PersonEmail,
		Username:    person.Username,
		RoleName:    person.RoleName,
		Permissions: formatPermissions(person.Permissions),
	})
}

// Logout handles POST /api/v1/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := h.sessionMgr.Get(r)
	if err != nil {
		h.logger.Debug("logout: session error", "error", err)
		JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	// Get person ID for logging before clearing
	personID, _ := h.sessionMgr.GetPersonID(session)

	h.sessionMgr.Clear(session)

	if err := h.sessionMgr.Save(r, w, session); err != nil {
		h.logger.Error("failed to clear session", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("user logged out", "person_id", personID)

	JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me handles GET /api/auth/me - returns current user from session (requires auth middleware)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	JSON(w, http.StatusOK, LoginResponse{
		PersonID:    person.PersonID,
		PersonName:  person.PersonName,
		PersonEmail: person.PersonEmail,
		Username:    person.Username,
		RoleName:    person.RoleName,
		Permissions: formatPermissions(person.Permissions),
	})
}

// UserPermissions handles GET /api/user/permissions (requires auth middleware)
func (h *AuthHandler) UserPermissions(w http.ResponseWriter, r *http.Request) {
	person := GetPersonFromContext(r.Context())
	if person == nil {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	JSON(w, http.StatusOK, UserPermissionsResponse{
		Permissions: formatPermissions(person.Permissions),
	})
}

// formatPermissions converts Permission slice to "resource:action" strings.
func formatPermissions(perms []domain.Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.Resource + ":" + p.Action
	}
	return result
}
