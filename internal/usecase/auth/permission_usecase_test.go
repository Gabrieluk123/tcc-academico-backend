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

type mockPermissionRepo struct{ mock.Mock }

func (m *mockPermissionRepo) Create(ctx context.Context, p *domain.Permission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPermissionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermissionRepo) FindBySlug(ctx context.Context, slug string) (*domain.Permission, error) {
	args := m.Called(ctx, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermissionRepo) List(ctx context.Context, params domain.PermissionListParams) ([]*domain.Permission, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockPermissionRepo) ListByRoleName(ctx context.Context, roleName string) ([]*domain.Permission, error) {
	args := m.Called(ctx, roleName)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ── ListPermissions ───────────────────────────────────────────────────────────

func TestListPermissions_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	perms := []*domain.Permission{
		{ID: uuid.New(), Slug: "user:create", Description: "Criar usuários"},
		{ID: uuid.New(), Slug: "user:read", Description: "Visualizar usuários"},
	}
	params := domain.PermissionListParams{}
	repo.On("List", ctx, params).Return(perms, 2, nil)

	result, total, err := auth.NewPermissionUseCase(repo).ListPermissions(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	assert.Equal(t, "user:create", result[0].Slug)
}

func TestListPermissions_Fail(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	params := domain.PermissionListParams{}
	repo.On("List", ctx, params).Return(nil, 0, errors.New("db error"))

	_, _, err := auth.NewPermissionUseCase(repo).ListPermissions(ctx, params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── FindPermissionByID ────────────────────────────────────────────────────────

func TestFindPermissionByID_Found(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	id := uuid.New()
	expected := &domain.Permission{ID: id, Slug: "user:create"}
	repo.On("FindByID", ctx, id).Return(expected, nil)

	result, err := auth.NewPermissionUseCase(repo).FindPermissionByID(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFindPermissionByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	id := uuid.New()
	repo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := auth.NewPermissionUseCase(repo).FindPermissionByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFindPermissionByID_Fail(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	id := uuid.New()
	repo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	_, err := auth.NewPermissionUseCase(repo).FindPermissionByID(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
