package domain

import (
	"context"
	"time"

	"errors"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("e-mail ou senha inválidos")

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
