package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        uuid.UUID
	Nome      string
	Descricao string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Permission struct {
	ID        uuid.UUID
	Slug      string
	Descricao string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           uuid.UUID
	Nome         string
	Email        string
	PasswordHash string
	RoleID       uuid.UUID
	Role         *Role

	IsActive   bool
	DisabledAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	List(ctx context.Context) ([]*User, error)
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
	ListUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type RoleUseCase interface {
	CreateRole(ctx context.Context, role *Role) error
	FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
}
