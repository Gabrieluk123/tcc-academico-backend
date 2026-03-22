package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Permission struct {
	ID          uuid.UUID
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	RoleID       uuid.UUID
	Role         *Role

	IsActive   bool
	DisabledAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserListParams struct {
	Page        int
	PageSize    int
	OrderBy     string
	OrderDir    string // "asc" ou "desc"
	RoleID      *uuid.UUID
	IsActive    *bool
	FirstName   *string
	LastName    *string
	Email       *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DisabledAt  *time.Time
	IncludeRole bool // quando true, carrega os detalhes da role associada
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, params UserListParams) ([]*User, int, error) // retorna lista, total, erro
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserUseCase interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, params UserListParams) ([]*User, int, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type RoleUseCase interface {
	CreateRole(ctx context.Context, role *Role) error
	FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type PermissionUseCase interface {
	ListPermissions(ctx context.Context) ([]*Permission, error)
	FindPermissionByID(ctx context.Context, id uuid.UUID) (*Permission, error)
}
