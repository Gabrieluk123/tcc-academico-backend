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

type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRoleRepo) List(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Role), args.Error(1)
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateRole(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{Nome: "Admin"}
	roleRepo.On("Create", ctx, role).Return(nil)

	uc := auth.NewRoleUseCase(roleRepo)

	err := uc.CreateRole(ctx, role)
	assert.NoError(t, err)
	roleRepo.AssertCalled(t, "Create", ctx, role)
}

func TestDeleteRole(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	roleRepo.On("Delete", ctx, uuid.Nil).Return(nil)

	uc := auth.NewRoleUseCase(roleRepo)

	err := uc.DeleteRole(ctx, uuid.Nil)
	assert.NoError(t, err)
	roleRepo.AssertCalled(t, "Delete", ctx, uuid.Nil)
}
