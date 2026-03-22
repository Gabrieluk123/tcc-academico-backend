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

// ── Mocks ────────────────────────────────────────────────────────────────────

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.User), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
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
	return m.Called(hash, password).Error(0)
}

type MockRefreshTokenRepository struct{ mock.Mock }

func (m *MockRefreshTokenRepository) Save(ctx context.Context, userID uuid.UUID, token string, expiresIn time.Duration) error {
	return m.Called(ctx, userID, token, expiresIn).Error(0)
}
func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockRefreshTokenRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	args := m.Called(ctx, token)
	if v := args.Get(0); v != nil {
		return v.(*domain.Session), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRefreshTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]domain.Session), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRefreshTokenRepository) RevokeByID(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}

// helper — avoids repetir sessionDuration em cada teste
func newAuthUC(
	userRepo *MockUserRepository,
	tokenGen *MockTokenGenerator,
	hash *MockHashProvider,
	repo *MockRefreshTokenRepository,
) domain.AuthUseCase {
	return auth.NewAuthUseCase(userRepo, tokenGen, hash, repo, 24*time.Hour)
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)
	refreshRepo := new(MockRefreshTokenRepository)

	user := &domain.User{ID: uuid.New(), Email: "john@email.com", PasswordHash: "hashed", IsActive: true}
	userRepo.On("FindByEmail", ctx, "john@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "password").Return(nil)
	tokenGen.On("GenerateToken", ctx, user).Return("token", nil)
	refreshRepo.On("Save", ctx, user.ID, "token", mock.Anything).Return(nil)

	token, err := newAuthUC(userRepo, tokenGen, hashProvider, refreshRepo).Login(ctx, "john@email.com", "password")

	assert.NoError(t, err)
	assert.Equal(t, "token", token)
}

func TestLogin_EmailNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)

	userRepo.On("FindByEmail", ctx, "notfound@email.com").Return(nil, errors.New("not found"))

	_, err := newAuthUC(userRepo, new(MockTokenGenerator), new(MockHashProvider), new(MockRefreshTokenRepository)).
		Login(ctx, "notfound@email.com", "password")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_UserInactive(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)

	user := &domain.User{ID: uuid.New(), Email: "inactive@email.com", PasswordHash: "hashed", IsActive: false}
	userRepo.On("FindByEmail", ctx, "inactive@email.com").Return(user, nil)

	_, err := newAuthUC(userRepo, new(MockTokenGenerator), new(MockHashProvider), new(MockRefreshTokenRepository)).
		Login(ctx, "inactive@email.com", "password")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)

	user := &domain.User{ID: uuid.New(), Email: "wrongpass@email.com", PasswordHash: "hashed", IsActive: true}
	userRepo.On("FindByEmail", ctx, "wrongpass@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "wrong").Return(errors.New("bcrypt mismatch"))

	_, err := newAuthUC(userRepo, new(MockTokenGenerator), hashProvider, new(MockRefreshTokenRepository)).
		Login(ctx, "wrongpass@email.com", "wrong")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_TokenGenerationFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)

	user := &domain.User{ID: uuid.New(), Email: "u@email.com", PasswordHash: "hashed", IsActive: true}
	userRepo.On("FindByEmail", ctx, "u@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "pass").Return(nil)
	tokenGen.On("GenerateToken", ctx, user).Return("", errors.New("signing error"))

	_, err := newAuthUC(userRepo, tokenGen, hashProvider, new(MockRefreshTokenRepository)).
		Login(ctx, "u@email.com", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing error")
}

func TestLogin_SaveSessionFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	hashProvider := new(MockHashProvider)
	tokenGen := new(MockTokenGenerator)
	refreshRepo := new(MockRefreshTokenRepository)

	user := &domain.User{ID: uuid.New(), Email: "u@email.com", PasswordHash: "hashed", IsActive: true}
	userRepo.On("FindByEmail", ctx, "u@email.com").Return(user, nil)
	hashProvider.On("CompareHash", "hashed", "pass").Return(nil)
	tokenGen.On("GenerateToken", ctx, user).Return("token", nil)
	refreshRepo.On("Save", ctx, user.ID, "token", mock.Anything).Return(errors.New("db error"))

	_, err := newAuthUC(userRepo, tokenGen, hashProvider, refreshRepo).Login(ctx, "u@email.com", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── Logout ───────────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	refreshRepo.On("Revoke", ctx, "token123").Return(nil)

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		Logout(ctx, "token123")

	assert.NoError(t, err)
}

func TestLogout_Fail(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	refreshRepo.On("Revoke", ctx, "token123").Return(errors.New("db error"))

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		Logout(ctx, "token123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── LogoutAll ────────────────────────────────────────────────────────────────

func TestLogoutAll_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()
	refreshRepo.On("RevokeAll", ctx, userID).Return(nil)

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		LogoutAll(ctx, userID)

	assert.NoError(t, err)
}

func TestLogoutAll_Fail(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()
	refreshRepo.On("RevokeAll", ctx, userID).Return(errors.New("db error"))

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		LogoutAll(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── ListSessions ─────────────────────────────────────────────────────────────

func TestListSessions_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()
	sessions := []domain.Session{
		{ID: "sess1", UserID: userID, Token: "token1"},
		{ID: "sess2", UserID: userID, Token: "token2"},
	}
	refreshRepo.On("ListByUser", ctx, userID).Return(sessions, nil)

	result, err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		ListSessions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "sess1", result[0].ID)
}

func TestListSessions_Fail(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()
	refreshRepo.On("ListByUser", ctx, userID).Return(nil, errors.New("db error"))

	_, err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		ListSessions(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── RevokeSession ─────────────────────────────────────────────────────────────

func TestRevokeSession_Success(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	refreshRepo.On("RevokeByID", ctx, "sess1").Return(nil)

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		RevokeSession(ctx, "sess1")

	assert.NoError(t, err)
}

func TestRevokeSession_Fail(t *testing.T) {
	ctx := context.Background()
	refreshRepo := new(MockRefreshTokenRepository)
	refreshRepo.On("RevokeByID", ctx, "sess1").Return(errors.New("db error"))

	err := newAuthUC(new(MockUserRepository), new(MockTokenGenerator), new(MockHashProvider), refreshRepo).
		RevokeSession(ctx, "sess1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
