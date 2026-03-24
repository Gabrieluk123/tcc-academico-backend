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

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.User), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) CountByRoleID(ctx context.Context, roleID uuid.UUID) (int, error) {
	args := m.Called(ctx, roleID)
	return args.Int(0), args.Error(1)
}
func (m *mockUserRepo) BulkUpdateRoleID(ctx context.Context, oldRoleID, newRoleID uuid.UUID) error {
	return m.Called(ctx, oldRoleID, newRoleID).Error(0)
}

type mockHashProvider struct{ mock.Mock }

func (m *mockHashProvider) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
func (m *mockHashProvider) CompareHash(hash, password string) error {
	return m.Called(hash, password).Error(0)
}

type mockPermRepoForUser struct{ mock.Mock }

func (m *mockPermRepoForUser) Create(ctx context.Context, p *domain.Permission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPermRepoForUser) FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForUser) FindBySlug(ctx context.Context, slug string) (*domain.Permission, error) {
	args := m.Called(ctx, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForUser) List(ctx context.Context, params domain.PermissionListParams) ([]*domain.Permission, int, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
func (m *mockPermRepoForUser) ListByRoleName(ctx context.Context, roleName string) ([]*domain.Permission, error) {
	args := m.Called(ctx, roleName)
	if v := args.Get(0); v != nil {
		return v.([]*domain.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPermRepoForUser) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ── CreateUser ────────────────────────────────────────────────────────────────

func TestCreateUser_HashesPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	user := &domain.User{FirstName: "Foo", LastName: "Bar", Email: "foo@bar.com", PasswordHash: "senha"}
	hashProv.On("HashPassword", "senha").Return("hashed", nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, "hashed", user.PasswordHash)
}

func TestCreateUser_HashFail(t *testing.T) {
	ctx := context.Background()
	hashProv := new(mockHashProvider)

	user := &domain.User{PasswordHash: "senha"}
	hashProv.On("HashPassword", "senha").Return("", errors.New("bcrypt error"))

	err := auth.NewUserUseCase(new(mockUserRepo), hashProv, new(mockPermRepoForUser)).CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bcrypt error")
}

func TestCreateUser_RepFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	user := &domain.User{PasswordHash: "senha"}
	hashProv.On("HashPassword", "senha").Return("hashed", nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── FindUserByID ──────────────────────────────────────────────────────────────

func TestFindUserByID_Found(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	expected := &domain.User{ID: id, FirstName: "Alice"}
	userRepo.On("FindByID", ctx, id).Return(expected, nil)

	result, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).FindUserByID(ctx, id, false)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFindUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	userRepo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).FindUserByID(ctx, id, false)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFindUserByID_Fail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	userRepo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	_, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).FindUserByID(ctx, id, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── FindUserByEmail ───────────────────────────────────────────────────────────

func TestFindUserByEmail_Found(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	expected := &domain.User{Email: "a@b.com"}
	userRepo.On("FindByEmail", ctx, "a@b.com").Return(expected, nil)

	result, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).FindUserByEmail(ctx, "a@b.com")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFindUserByEmail_Fail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	userRepo.On("FindByEmail", ctx, "a@b.com").Return(nil, errors.New("db error"))

	_, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).FindUserByEmail(ctx, "a@b.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── UpdateUser ────────────────────────────────────────────────────────────────

func TestUpdateUser_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	user := &domain.User{ID: uuid.New(), FirstName: "Updated"}
	userRepo.On("Update", ctx, user).Return(nil)

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).UpdateUser(ctx, user)

	assert.NoError(t, err)
}

func TestUpdateUser_Fail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	user := &domain.User{ID: uuid.New()}
	userRepo.On("Update", ctx, user).Return(errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).UpdateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── DeleteUser ────────────────────────────────────────────────────────────────

func TestDeleteUser_LogsAndDisables(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	user := &domain.User{ID: uuid.Nil, FirstName: "Foo", LastName: "Bar", IsActive: true}
	userRepo.On("FindByID", ctx, user.ID).Return(user, nil)
	userRepo.On("Update", ctx, user).Return(nil)

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).DeleteUser(ctx, user.ID)

	assert.NoError(t, err)
	assert.False(t, user.IsActive)
	assert.NotNil(t, user.DisabledAt)
}

func TestDeleteUser_FindFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	userRepo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).DeleteUser(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestDeleteUser_UpdateFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	user := &domain.User{ID: uuid.New(), IsActive: true}
	userRepo.On("FindByID", ctx, user.ID).Return(user, nil)
	userRepo.On("Update", ctx, user).Return(errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).DeleteUser(ctx, user.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── ListUsers ─────────────────────────────────────────────────────────────────

func TestListUsers_PaginationAndFilters(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	params := domain.UserListParams{Page: 1, PageSize: 2, OrderBy: "first_name", OrderDir: "asc"}
	users := []*domain.User{
		{ID: uuid.New(), FirstName: "Alice"},
		{ID: uuid.New(), FirstName: "Bob"},
	}
	userRepo.On("List", ctx, params).Return(users, 2, nil)

	result, count, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).ListUsers(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, result, 2)
	assert.Equal(t, "Alice", result[0].FirstName)
}

func TestListUsers_FilterByRole(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	roleID := uuid.New()
	params := domain.UserListParams{Page: 1, PageSize: 10, OrderBy: "email", OrderDir: "desc", RoleID: &roleID}
	users := []*domain.User{{ID: uuid.New(), RoleID: roleID}}
	userRepo.On("List", ctx, params).Return(users, 1, nil)

	result, count, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).ListUsers(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, roleID, result[0].RoleID)
}

func TestListUsers_Fail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	params := domain.UserListParams{Page: 1, PageSize: 10}
	userRepo.On("List", ctx, params).Return(nil, 0, errors.New("db error"))

	_, _, err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).ListUsers(ctx, params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ── ChangePassword ────────────────────────────────────────────────────────────

func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	id := uuid.New()
	user := &domain.User{ID: id, PasswordHash: "old_hash", IsActive: true}
	userRepo.On("FindByID", ctx, id).Return(user, nil)
	hashProv.On("CompareHash", "old_hash", "old_pass").Return(nil)
	hashProv.On("HashPassword", "new_pass").Return("new_hash", nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).ChangePassword(ctx, id, "old_pass", "new_pass")

	assert.NoError(t, err)
	assert.Equal(t, "new_hash", user.PasswordHash)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	id := uuid.New()
	user := &domain.User{ID: id, PasswordHash: "old_hash"}
	userRepo.On("FindByID", ctx, id).Return(user, nil)
	hashProv.On("CompareHash", "old_hash", "wrong_pass").Return(errors.New("mismatch"))

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).ChangePassword(ctx, id, "wrong_pass", "new_pass")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestChangePassword_UserNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	userRepo.On("FindByID", ctx, id).Return(nil, nil)

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).ChangePassword(ctx, id, "old", "new")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestChangePassword_FindFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)

	id := uuid.New()
	userRepo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, new(mockHashProvider), new(mockPermRepoForUser)).ChangePassword(ctx, id, "old", "new")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestChangePassword_HashFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	id := uuid.New()
	user := &domain.User{ID: id, PasswordHash: "old_hash"}
	userRepo.On("FindByID", ctx, id).Return(user, nil)
	hashProv.On("CompareHash", "old_hash", "old_pass").Return(nil)
	hashProv.On("HashPassword", "new_pass").Return("", errors.New("bcrypt error"))

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).ChangePassword(ctx, id, "old_pass", "new_pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bcrypt error")
}

func TestChangePassword_UpdateFail(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	hashProv := new(mockHashProvider)

	id := uuid.New()
	user := &domain.User{ID: id, PasswordHash: "old_hash"}
	userRepo.On("FindByID", ctx, id).Return(user, nil)
	hashProv.On("CompareHash", "old_hash", "old_pass").Return(nil)
	hashProv.On("HashPassword", "new_pass").Return("new_hash", nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

	err := auth.NewUserUseCase(userRepo, hashProv, new(mockPermRepoForUser)).ChangePassword(ctx, id, "old_pass", "new_pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
