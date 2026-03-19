package auth_test

import (
	"context"
	"testing"

	"academico/internal/domain"
	"academico/internal/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para UserRepository
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *mockUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.User), args.Error(1)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock para HashProvider
type mockHashProvider struct {
	mock.Mock
}

func (m *mockHashProvider) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
func (m *mockHashProvider) CompareHash(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}

func TestCreateUser_HashesPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	user := &domain.User{FirstName: "Foo", LastName: "Bar", Email: "foo@bar.com", PasswordHash: "senha"}

	hashProv.On("HashPassword", "senha").Return("hashed", nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := auth.NewUserUseCase(userRepo, hashProv)

	err := uc.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, "hashed", user.PasswordHash)
	userRepo.AssertCalled(t, "Create", ctx, user)
}

func TestDeleteUser_LogsAndDisables(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	user := &domain.User{ID: uuid.Nil, FirstName: "Foo", LastName: "Bar", IsActive: true}
	userRepo.On("FindByID", ctx, user.ID).Return(user, nil)
	userRepo.On("Update", ctx, user).Return(nil)

	uc := auth.NewUserUseCase(userRepo, hashProv)

	err := uc.DeleteUser(ctx, user.ID)
	assert.NoError(t, err)
	assert.False(t, user.IsActive)
	assert.NotNil(t, user.DisabledAt)
	userRepo.AssertCalled(t, "Update", ctx, user)
}
