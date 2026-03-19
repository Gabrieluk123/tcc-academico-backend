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

func (m *MockUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
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
