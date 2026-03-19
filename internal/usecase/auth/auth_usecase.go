package auth

import (
	"context"
	"fmt"
	"strings"

	"academico/internal/domain"
	"log/slog"
)

var _ domain.AuthUseCase = (*authUseCase)(nil)

type authUseCase struct {
	userRepo     domain.UserRepository
	tokenGen     domain.TokenGenerator
	hashProvider domain.HashProvider
	refreshRepo  domain.RefreshTokenRepository
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenGen domain.TokenGenerator,
	hashProvider domain.HashProvider,
	refreshRepo domain.RefreshTokenRepository,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		tokenGen:     tokenGen,
		hashProvider: hashProvider,
		refreshRepo:  refreshRepo,
	}
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		masked := maskEmail(email)
		slog.ErrorContext(ctx, "falha ao buscar usuario por email", slog.String("email", masked), slog.String("error", err.Error()))
		return "", fmt.Errorf("email nao encontrado: %w", err)
	}

	if !user.IsActive {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "usuario inativo", slog.String("email", masked), slog.String("user_id", user.ID.String()))
		return "", fmt.Errorf("usuario inativo: %w", err)
	}

	err = uc.hashProvider.CompareHash(user.PasswordHash, password)
	if err != nil {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "senha incorreta", slog.String("email", masked), slog.String("user_id", user.ID.String()))
		return "", fmt.Errorf("senha incorreta: %w", err)
	}

	token, err := uc.tokenGen.GenerateToken(ctx, user)
	if err != nil {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "falha ao gerar token", slog.String("email", masked), slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("erro ao gerar token: %w", err)
	}

	err = uc.refreshRepo.Save(ctx, user.ID, token, 0)
	if err != nil {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "falha ao salvar refresh token", slog.String("email", masked), slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("erro ao salvar refresh token: %w", err)
	}

	return token, nil
}

// maskEmail mascara o email para logs (ex: j***@email.com)
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) < 1 {
		return "***"
	}
	return parts[0][:1] + "***@" + parts[1]
}
