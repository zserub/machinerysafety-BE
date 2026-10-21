package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zserub/machinerysafety-BE/internal/domain"
	"github.com/zserub/machinerysafety-BE/internal/service"
	"github.com/zserub/machinerysafety-BE/internal/utils"
)

func strPtr(s string) *string {
	return &s
}

// MockPersonRepo is a mock implementation of PersonRepository.
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

func TestAuthService_Login(t *testing.T) {
	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Generate a valid password hash for testing
	validPassword := "testpassword123"
	validHash, _ := utils.HashPassword(validPassword, 10)

	tests := []struct {
		name       string
		username   string
		password   string
		setupMock  func(*MockPersonRepo)
		wantErr    error
		wantPerson bool
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: validPassword,
			setupMock: func(m *MockPersonRepo) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&domain.Person{
					PersonID:   1,
					PersonName: strPtr("Test User"),
					Username:   strPtr("testuser"),
					PwHash:     &validHash,
					Locked:     false,
					RoleName:   "admin",
					Permissions: []domain.Permission{
						{Resource: "risks", Action: "read"},
					},
				}, nil)
			},
			wantErr:    nil,
			wantPerson: true,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password",
			setupMock: func(m *MockPersonRepo) {
				m.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, domain.ErrPersonNotFound)
			},
			wantErr:    domain.ErrInvalidCredentials,
			wantPerson: false,
		},
		{
			name:     "account locked",
			username: "lockeduser",
			password: validPassword,
			setupMock: func(m *MockPersonRepo) {
				m.On("GetByUsername", mock.Anything, "lockeduser").Return(&domain.Person{
					PersonID: 2,
					Username: strPtr("lockeduser"),
					PwHash:   &validHash,
					Locked:   true,
				}, nil)
			},
			wantErr:    domain.ErrAccountLocked,
			wantPerson: false,
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(m *MockPersonRepo) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&domain.Person{
					PersonID: 1,
					Username: strPtr("testuser"),
					PwHash:   &validHash,
					Locked:   false,
				}, nil)
			},
			wantErr:    domain.ErrInvalidCredentials,
			wantPerson: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPersonRepo)
			tt.setupMock(mockRepo)

			svc := service.NewAuthService(mockRepo, logger)
			ctx := context.Background()

			person, err := svc.Login(ctx, tt.username, tt.password)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, person)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, person)
				assert.Nil(t, person.PwHash) // Password hash should be cleared
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
