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
func (m *mockRoleRepo) List(ctx context.Context, params domain.RoleListParams) ([]*domain.Role, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Role), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// mockPermRepoForRole satisfies domain.PermissionRepository for role usecase tests.
type mockPermRepoForRole struct{ mock.Mock }

func (m *mockPermRepoForRole) Create(ctx context.Context, p *domain.Permission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPermRepoForRole) FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForRole) FindBySlug(ctx context.Context, slug string) (*domain.Permission, error) {
	args := m.Called(ctx, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForRole) List(ctx context.Context, params domain.PermissionListParams) ([]*domain.Permission, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockPermRepoForRole) ListByRoleName(ctx context.Context, roleName string) ([]*domain.Permission, error) {
	args := m.Called(ctx, roleName)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForRole) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockUserRepoForRole struct{ mock.Mock }

func (m *mockUserRepoForRole) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepoForRole) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepoForRole) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepoForRole) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepoForRole) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.User), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockUserRepoForRole) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepoForRole) CountByRoleID(ctx context.Context, roleID uuid.UUID) (int, error) {
	args := m.Called(ctx, roleID)
	return args.Int(0), args.Error(1)
}
func (m *mockUserRepoForRole) BulkUpdateRoleID(ctx context.Context, oldRoleID, newRoleID uuid.UUID) error {
	return m.Called(ctx, oldRoleID, newRoleID).Error(0)
}

func newRoleUC(roleRepo *mockRoleRepo) domain.RoleUseCase {
	return auth.NewRoleUseCase(roleRepo, new(mockPermRepoForRole), new(mockUserRepoForRole))
}

func newRoleUCWithPerm(roleRepo *mockRoleRepo, permRepo *mockPermRepoForRole) domain.RoleUseCase {
	return auth.NewRoleUseCase(roleRepo, permRepo, new(mockUserRepoForRole))
}

func newRoleUCWithUser(roleRepo *mockRoleRepo, userRepo *mockUserRepoForRole) domain.RoleUseCase {
	return auth.NewRoleUseCase(roleRepo, new(mockPermRepoForRole), userRepo)
}

// ── CreateRole ────────────────────────────────────────────────────────────────

func TestCreateRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{Name: "Admin"}
	roleRepo.On("Create", ctx, role).Return(nil)

	err := newRoleUC(roleRepo).CreateRole(ctx, role)

	assert.NoError(t, err)
}

func TestCreateRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{Name: "Admin"}
	roleRepo.On("Create", ctx, role).Return(errors.New("db error"))

	err := newRoleUC(roleRepo).CreateRole(ctx, role)

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

	result, err := newRoleUC(roleRepo).FindRoleByID(ctx, id, false)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFindRoleByID_WithPermissions(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)
	permRepo := new(mockPermRepoForRole)

	id := uuid.New()
	role := &domain.Role{ID: id, Name: "Admin"}
	perms := []*domain.Permission{{ID: uuid.New(), Slug: "user:create"}}

	roleRepo.On("FindByID", ctx, id).Return(role, nil)
	permRepo.On("ListByRoleName", ctx, "Admin").Return(perms, nil)

	result, err := newRoleUCWithPerm(roleRepo, permRepo).FindRoleByID(ctx, id, true)

	assert.NoError(t, err)
	assert.Equal(t, perms, result.Permissions)
}

func TestFindRoleByID_NotFound(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	roleRepo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := newRoleUC(roleRepo).FindRoleByID(ctx, id, false)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFindRoleByID_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	id := uuid.New()
	roleRepo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	_, err := newRoleUC(roleRepo).FindRoleByID(ctx, id, false)

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
	params := domain.RoleListParams{}
	roleRepo.On("List", ctx, params).Return(roles, 2, nil)

	result, total, err := newRoleUC(roleRepo).ListRoles(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	assert.Equal(t, "Admin", result[0].Name)
}

func TestListRoles_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	params := domain.RoleListParams{}
	roleRepo.On("List", ctx, params).Return(nil, 0, errors.New("db error"))

	_, _, err := newRoleUC(roleRepo).ListRoles(ctx, params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── UpdateRole ────────────────────────────────────────────────────────────────

func TestUpdateRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{ID: uuid.New(), Name: "Updated"}
	roleRepo.On("Update", ctx, role).Return(nil)

	err := newRoleUC(roleRepo).UpdateRole(ctx, role)

	assert.NoError(t, err)
}

func TestUpdateRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)

	role := &domain.Role{ID: uuid.New()}
	roleRepo.On("Update", ctx, role).Return(errors.New("db error"))

	err := newRoleUC(roleRepo).UpdateRole(ctx, role)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── DeleteRole ────────────────────────────────────────────────────────────────

func TestDeleteRole_Success(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)
	userRepo := new(mockUserRepoForRole)

	userRepo.On("CountByRoleID", ctx, uuid.Nil).Return(0, nil)
	roleRepo.On("Delete", ctx, uuid.Nil).Return(nil)

	err := newRoleUCWithUser(roleRepo, userRepo).DeleteRole(ctx, uuid.Nil)

	assert.NoError(t, err)
}

func TestDeleteRole_Fail(t *testing.T) {
	ctx := context.Background()
	roleRepo := new(mockRoleRepo)
	userRepo := new(mockUserRepoForRole)

	id := uuid.New()
	userRepo.On("CountByRoleID", ctx, id).Return(0, nil)
	roleRepo.On("Delete", ctx, id).Return(errors.New("db error"))

	err := newRoleUCWithUser(roleRepo, userRepo).DeleteRole(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestDeleteRole_BlockedWhenUsersExist(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepoForRole)

	id := uuid.New()
	userRepo.On("CountByRoleID", ctx, id).Return(3, nil)

	err := newRoleUCWithUser(new(mockRoleRepo), userRepo).DeleteRole(ctx, id)

	assert.ErrorIs(t, err, domain.ErrRoleInUse)
}

// ── ReassignUsers ─────────────────────────────────────────────────────────────

func TestReassignUsers_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepoForRole)

	fromID := uuid.New()
	toID := uuid.New()
	userRepo.On("CountByRoleID", ctx, fromID).Return(5, nil)
	userRepo.On("BulkUpdateRoleID", ctx, fromID, toID).Return(nil)

	err := newRoleUCWithUser(new(mockRoleRepo), userRepo).ReassignUsers(ctx, fromID, toID)

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "BulkUpdateRoleID", ctx, fromID, toID)
}

func TestReassignUsers_NoUsersIsNoop(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepoForRole)

	fromID := uuid.New()
	toID := uuid.New()
	userRepo.On("CountByRoleID", ctx, fromID).Return(0, nil)

	err := newRoleUCWithUser(new(mockRoleRepo), userRepo).ReassignUsers(ctx, fromID, toID)

	assert.NoError(t, err)
	userRepo.AssertNotCalled(t, "BulkUpdateRoleID")
}

func TestReassignUsers_BulkFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepoForRole)

	fromID := uuid.New()
	toID := uuid.New()
	userRepo.On("CountByRoleID", ctx, fromID).Return(2, nil)
	userRepo.On("BulkUpdateRoleID", ctx, fromID, toID).Return(errors.New("db error"))

	err := newRoleUCWithUser(new(mockRoleRepo), userRepo).ReassignUsers(ctx, fromID, toID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
