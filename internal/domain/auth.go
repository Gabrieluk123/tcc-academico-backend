package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("e-mail ou senha inválidos")
	ErrForbidden          = errors.New("acesso negado ao recurso")
)

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type TokenGenerator interface {
	GenerateToken(ctx context.Context, user *User) (string, error)
}

type HashProvider interface {
	HashPassword(password string) (string, error)
	CompareHash(hash, password string) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, userID uuid.UUID, token string, expiresIn time.Duration) error
	Revoke(ctx context.Context, token string) error
}

// Enforcer abstracts the authorization engine (Casbin).
// sub = role name, obj = resource (e.g. "user", "diary"), act = action (e.g. "create", "read", "*")
type Enforcer interface {
	Enforce(ctx context.Context, roleName, resource, action string) (bool, error)
	AddPolicy(ctx context.Context, roleName, resource, action string) (bool, error)
	RemovePolicy(ctx context.Context, roleName, resource, action string) (bool, error)
}

// PermissionRepository is the catalog of available permissions that can be
// granted to roles via the Enforcer.
type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	FindBySlug(ctx context.Context, slug string) (*Permission, error)
	List(ctx context.Context) ([]*Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
