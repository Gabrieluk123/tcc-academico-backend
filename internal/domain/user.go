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

	// Controle de Acesso e Regra de Negócio
	IsActive   bool       // true = pode logar; false = acesso bloqueado
	DisabledAt *time.Time // Ponteiro (*). Se for nil, significa que nunca foi desativado (ou está ativo)

	// Auditoria padrão
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
