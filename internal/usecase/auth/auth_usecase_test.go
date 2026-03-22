package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"academico/internal/domain"
	"academico/internal/usecase/auth"
)

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockTokenGenerator struct{ mock.Mock }

func (m *MockTokenGenerator) GenerateToken(ctx context.Context, user *domain.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

type MockHashProvider struct{ mock.Mock }

func (m *MockHashProvider) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
func (m *MockHashProvider) CompareHash(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}

type MockRefreshTokenRepository struct{ mock.Mock }

func (m *MockRefreshTokenRepository) Save(ctx context.Context, userID uuid.UUID, token string, expiresIn time.Duration) error {
	args := m.Called(ctx, userID, token, expiresIn)
	return args.Error(0)
}
func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
func (m *MockRefreshTokenRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockRefreshTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Session), args.Error(1)
}
func (m *MockRefreshTokenRepository) RevokeByID(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}
func TestLogoutAll_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	userID := uuid.New()
	refreshRepo.On("RevokeAll", ctx, userID).Return(nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	err := uc.LogoutAll(ctx, userID)

	assert.NoError(t, err)
	refreshRepo.AssertCalled(t, "RevokeAll", ctx, userID)
}

func TestListSessions_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	userID := uuid.New()
	sessions := []domain.Session{
		{ID: "sess1", UserID: userID, Token: "token1", Revoked: false},
		{ID: "sess2", UserID: userID, Token: "token2", Revoked: false},
	}
	refreshRepo.On("ListByUser", ctx, userID).Return(sessions, nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	result, err := uc.ListSessions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "sess1", result[0].ID)
	refreshRepo.AssertCalled(t, "ListByUser", ctx, userID)
}

func TestRevokeSession_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	sessionID := "sess1"
	refreshRepo.On("RevokeByID", ctx, sessionID).Return(nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	err := uc.RevokeSession(ctx, sessionID)

	assert.NoError(t, err)
	refreshRepo.AssertCalled(t, "RevokeByID", ctx, sessionID)
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)
	refreshRepo := new(MockRefreshTokenRepository)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "john@email.com",
		PasswordHash: "hashed",
		IsActive:     true,
	}

	userRepo.On("FindByEmail", ctx, "john@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "password").Return(nil)
	tokenGen.On("GenerateToken", ctx, user).Return("token", nil)
	refreshRepo.On("Save", ctx, user.ID, "token", mock.Anything).Return(nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	token, err := uc.Login(ctx, "john@email.com", "password")

	assert.NoError(t, err)
	assert.Equal(t, "token", token)
}

func TestLogin_EmailNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	refreshRepo := new(MockRefreshTokenRepository)
	tokenGen := new(MockTokenGenerator)

	userRepo.On("FindByEmail", ctx, "notfound@email.com").Return((*domain.User)(nil), errors.New("registro não encontrado no bd"))

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	_, err := uc.Login(ctx, "notfound@email.com", "password")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_UserInactive(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	refreshRepo := new(MockRefreshTokenRepository)
	tokenGen := new(MockTokenGenerator)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "inactive@email.com",
		PasswordHash: "hashed",
		IsActive:     false,
	}

	userRepo.On("FindByEmail", ctx, "inactive@email.com").Return(user, nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	_, err := uc.Login(ctx, "inactive@email.com", "password")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inativo")
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	refreshRepo := new(MockRefreshTokenRepository)
	tokenGen := new(MockTokenGenerator)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "wrongpass@email.com",
		PasswordHash: "hashed",
		IsActive:     true,
	}

	userRepo.On("FindByEmail", ctx, "wrongpass@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "wrong").Return(errors.New("bcrypt hash mismatch"))

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	_, err := uc.Login(ctx, "wrongpass@email.com", "wrong")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	refreshRepo.On("Revoke", ctx, "token123").Return(nil)

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	err := uc.Logout(ctx, "token123")

	assert.NoError(t, err)
	refreshRepo.AssertCalled(t, "Revoke", ctx, "token123")
}

func TestLogout_Fail(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	refreshRepo.On("Revoke", ctx, "token123").Return(errors.New("db error"))

	uc := auth.NewAuthUseCase(userRepo, tokenGen, hashProvider, refreshRepo)
	err := uc.Logout(ctx, "token123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	refreshRepo.AssertCalled(t, "Revoke", ctx, "token123")
}
