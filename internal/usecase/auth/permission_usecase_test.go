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
func (m *mockPermissionRepo) List(ctx context.Context) ([]*domain.Permission, error) {
	args := m.Called(ctx)
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
	repo.On("List", ctx).Return(perms, nil)

	result, err := auth.NewPermissionUseCase(repo).ListPermissions(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user:create", result[0].Slug)
}

func TestListPermissions_Fail(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPermissionRepo)

	repo.On("List", ctx).Return(nil, errors.New("db error"))

	_, err := auth.NewPermissionUseCase(repo).ListPermissions(ctx)

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
