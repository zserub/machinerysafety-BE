package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zserub/machinerysafety-BE/internal/config"
	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/handler"
)

// MockAuthService is a mock implementation of AuthServiceInterface.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*domain.Person, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func newTestSessionManager() *handler.SessionManager {
	cfg := &config.Config{
		SessionKey:      "test-key-must-be-32-bytes-long!!", // 32 bytes
		SessionName:     "test_session",
		SessionMaxAge:   3600,
		SessionHTTPOnly: true,
		SessionSecure:   false,
		SessionSameSite: "lax",
	}
	return handler.NewSessionManager(cfg)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func strPtr(s string) *string {
	return &s
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		wantStatusCode int
		wantError      bool
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "password123").Return(&domain.Person{
					PersonID:    1,
					PersonName:  strPtr("Test User"),
					PersonEmail: strPtr("test@example.com"),
					Username:    strPtr("testuser"),
					RoleName:    "admin",
					Permissions: []domain.Permission{
						{Resource: "entries", Action: "view"},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "wrongpassword").Return(nil, domain.ErrInvalidCredentials)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantError:      true,
		},
		{
			name: "account locked",
			requestBody: map[string]string{
				"username": "lockeduser",
				"password": "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "lockeduser", "password123").Return(nil, domain.ErrAccountLocked)
			},
			wantStatusCode: http.StatusForbidden,
			wantError:      true,
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
			},
			setupMock:      func(m *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid JSON",
			requestBody:    "not valid json",
			setupMock:      func(m *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.setupMock(mockService)

			sessionMgr := newTestSessionManager()
			logger := newTestLogger()

			h := handler.NewAuthHandler(mockService, sessionMgr, logger)

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			var response map[string]interface{}
			json.Unmarshal(rr.Body.Bytes(), &response)

			if tt.wantError {
				assert.Contains(t, response, "error")
			} else {
				assert.Contains(t, response, "person_id")
				assert.Contains(t, response, "username")
				// Check that session cookie was set
				cookies := rr.Result().Cookies()
				assert.NotEmpty(t, cookies)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	mockService := new(MockAuthService)
	sessionMgr := newTestSessionManager()
	logger := newTestLogger()

	h := handler.NewAuthHandler(mockService, sessionMgr, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "logged out", response["message"])
}

func TestAuthHandler_UserPermissions(t *testing.T) {
	tests := []struct {
		name           string
		person         *domain.Person
		wantStatusCode int
	}{
		{
			name: "authenticated user",
			person: &domain.Person{
				PersonID:    1,
				PersonName:  strPtr("Test User"),
				PersonEmail: strPtr("test@example.com"),
				Username:    strPtr("testuser"),
				RoleName:    "admin",
				Permissions: []domain.Permission{
					{Resource: "entries", Action: "view"},
					{Resource: "dashboard", Action: "view"},
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "no user in context",
			person:         nil,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			sessionMgr := newTestSessionManager()
			logger := newTestLogger()

			h := handler.NewAuthHandler(mockService, sessionMgr, logger)

			req := httptest.NewRequest(http.MethodGet, "/api/user/permissions", nil)

			if tt.person != nil {
				ctx := context.WithValue(req.Context(), handler.ContextKeyPerson, tt.person)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			h.UserPermissions(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantStatusCode == http.StatusOK {
				var response handler.UserPermissionsResponse
				json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Len(t, response.Permissions, len(tt.person.Permissions))
				assert.Contains(t, response.Permissions, "entries:view")
				assert.Contains(t, response.Permissions, "dashboard:view")
			}
		})
	}
}
