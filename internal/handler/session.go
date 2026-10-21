package handler

import (
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/zserub/machinerysafety-BE/internal/config"
)

const (
	sessionKeyPersonID = "person_id"
	sessionKeyUsername = "username"
)

// SessionManager wraps gorilla/sessions.CookieStore.
type SessionManager struct {
	store       *sessions.CookieStore
	sessionName string
}

// NewSessionManager creates a new session manager.
func NewSessionManager(cfg *config.Config) *SessionManager {
	store := sessions.NewCookieStore([]byte(cfg.SessionKey))

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.SessionMaxAge,
		HttpOnly: cfg.SessionHTTPOnly,
		Secure:   cfg.SessionSecure,
		SameSite: parseSameSite(cfg.SessionSameSite),
	}

	return &SessionManager{
		store:       store,
		sessionName: cfg.SessionName,
	}
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// Get retrieves the session from the request.
func (sm *SessionManager) Get(r *http.Request) (*sessions.Session, error) {
	return sm.store.Get(r, sm.sessionName)
}

// Save saves the session.
func (sm *SessionManager) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return session.Save(r, w)
}

// SetPersonID stores the person ID in the session.
func (sm *SessionManager) SetPersonID(session *sessions.Session, personID int16) {
	session.Values[sessionKeyPersonID] = personID
}

// GetPersonID retrieves the person ID from the session.
func (sm *SessionManager) GetPersonID(session *sessions.Session) (int16, bool) {
	val, ok := session.Values[sessionKeyPersonID]
	if !ok {
		return 0, false
	}
	personID, ok := val.(int16)
	return personID, ok
}

// SetUsername stores the username in the session.
func (sm *SessionManager) SetUsername(session *sessions.Session, username string) {
	session.Values[sessionKeyUsername] = username
}

// GetUsername retrieves the username from the session.
func (sm *SessionManager) GetUsername(session *sessions.Session) (string, bool) {
	val, ok := session.Values[sessionKeyUsername]
	if !ok {
		return "", false
	}
	username, ok := val.(string)
	return username, ok
}

// Clear clears the session values and marks it for deletion.
func (sm *SessionManager) Clear(session *sessions.Session) {
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
}
