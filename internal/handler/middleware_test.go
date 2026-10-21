package handler_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/handler"
)

func ptrStr(s string) *string {
	return &s
}

// MockPersonRepo for middleware tests
type MockPersonRepo struct {
	mock.Mock
}

func (m *MockPersonRepo) GetByUsername(ctx context.Context, username string) (*domain.Person, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func (m *MockPersonRepo) GetByID(ctx context.Context, id int16) (*domain.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func TestRequirePermission(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockRepo := new(MockPersonRepo)
	sessionMgr := newTestSessionManager()

	authMiddleware := handler.NewAuthMiddleware(sessionMgr, mockRepo, logger)

	tests := []struct {
		name           string
		person         *domain.Person
		resource       string
		action         string
		wantStatusCode int
	}{
		{
			name: "has permission",
			person: &domain.Person{
				PersonID: 1,
				Permissions: []domain.Permission{
					{Resource: "entries", Action: "view"},
					{Resource: "entries", Action: "create"},
				},
			},
			resource:       "entries",
			action:         "view",
			wantStatusCode: http.StatusOK,
		},
		{
			name: "lacks permission",
			person: &domain.Person{
				PersonID: 1,
				Permissions: []domain.Permission{
					{Resource: "entries", Action: "view"},
				},
			},
			resource:       "entries",
			action:         "delete",
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "no person in context",
			person:         nil,
			resource:       "entries",
			action:         "view",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that returns 200
			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := authMiddleware.RequirePermission(tt.resource, tt.action)
			wrappedHandler := middleware(innerHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.person != nil {
				ctx := context.WithValue(req.Context(), handler.ContextKeyPerson, tt.person)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestGetPersonFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   *domain.Person
		wantOK bool
	}{
		{
			name: "person exists in context",
			ctx: context.WithValue(context.Background(), handler.ContextKeyPerson, &domain.Person{
				PersonID: 1,
				Username: ptrStr("testuser"),
			}),
			want: &domain.Person{
				PersonID: 1,
				Username: ptrStr("testuser"),
			},
			wantOK: true,
		},
		{
			name:   "no person in context",
			ctx:    context.Background(),
			want:   nil,
			wantOK: false,
		},
		{
			name:   "wrong type in context",
			ctx:    context.WithValue(context.Background(), handler.ContextKeyPerson, "not a person"),
			want:   nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.GetPersonFromContext(tt.ctx)

			if tt.wantOK {
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.PersonID, got.PersonID)
				assert.Equal(t, tt.want.Username, got.Username)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}
