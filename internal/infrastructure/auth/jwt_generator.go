package auth

import (
	"context"
	"time"

	"academico/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// Garante em tempo de compilação que jwtGenerator implementa domain.TokenGenerator
var _ domain.TokenGenerator = (*jwtGenerator)(nil)

type jwtGenerator struct {
	secret    []byte
	expiresIn time.Duration
}

func NewJWTGenerator(secret string, expiresIn time.Duration) domain.TokenGenerator {
	return &jwtGenerator{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func (j *jwtGenerator) GenerateToken(ctx context.Context, user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role_id": user.RoleID.String(),
		"exp":     time.Now().Add(j.expiresIn).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}
