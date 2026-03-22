package auth_test

import (
	"context"
	"errors"
	"testing"

	"academico/internal/domain"
	"academico/internal/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.Role), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRoleRepo) List(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Role), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ── CreateRole ────────────────────────────────────────────────────────────────

func TestCreateRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{Name: "Admin"}
	roleRepo.On("Create", ctx, role).Return(nil)

	err := auth.NewRoleUseCase(roleRepo).CreateRole(ctx, role)

	assert.NoError(t, err)
}

func TestCreateRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{Name: "Admin"}
	roleRepo.On("Create", ctx, role).Return(errors.New("db error"))

	err := auth.NewRoleUseCase(roleRepo).CreateRole(ctx, role)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── FindRoleByID ──────────────────────────────────────────────────────────────

func TestFindRoleByID_Found(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	expected := &domain.Role{ID: id, Name: "Professor"}
	roleRepo.On("FindByID", ctx, id).Return(expected, nil)

	result, err := auth.NewRoleUseCase(roleRepo).FindRoleByID(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFindRoleByID_NotFound(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	roleRepo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := auth.NewRoleUseCase(roleRepo).FindRoleByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFindRoleByID_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	roleRepo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	_, err := auth.NewRoleUseCase(roleRepo).FindRoleByID(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── ListRoles ─────────────────────────────────────────────────────────────────

func TestListRoles_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	roles := []*domain.Role{
		{ID: uuid.New(), Name: "Admin"},
		{ID: uuid.New(), Name: "Professor"},
	}
	roleRepo.On("List", ctx).Return(roles, nil)

	result, err := auth.NewRoleUseCase(roleRepo).ListRoles(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Admin", result[0].Name)
}

func TestListRoles_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	roleRepo.On("List", ctx).Return(nil, errors.New("db error"))

	_, err := auth.NewRoleUseCase(roleRepo).ListRoles(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── UpdateRole ────────────────────────────────────────────────────────────────

func TestUpdateRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{ID: uuid.New(), Name: "Updated"}
	roleRepo.On("Update", ctx, role).Return(nil)

	err := auth.NewRoleUseCase(roleRepo).UpdateRole(ctx, role)

	assert.NoError(t, err)
}

func TestUpdateRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{ID: uuid.New()}
	roleRepo.On("Update", ctx, role).Return(errors.New("db error"))

	err := auth.NewRoleUseCase(roleRepo).UpdateRole(ctx, role)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── DeleteRole ────────────────────────────────────────────────────────────────

func TestDeleteRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	roleRepo.On("Delete", ctx, uuid.Nil).Return(nil)

	err := auth.NewRoleUseCase(roleRepo).DeleteRole(ctx, uuid.Nil)

	assert.NoError(t, err)
}

func TestDeleteRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	roleRepo.On("Delete", ctx, id).Return(errors.New("db error"))

	err := auth.NewRoleUseCase(roleRepo).DeleteRole(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
