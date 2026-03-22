package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"academico/internal/domain"
	"log/slog"

	"github.com/google/uuid"
)

var _ domain.AuthUseCase = (*authUseCase)(nil)

type authUseCase struct {
	userRepo        domain.UserRepository
	tokenGen        domain.TokenGenerator
	hashProvider    domain.HashProvider
	refreshRepo     domain.RefreshTokenRepository
	sessionDuration time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenGen domain.TokenGenerator,
	hashProvider domain.HashProvider,
	refreshRepo domain.RefreshTokenRepository,
	sessionDuration time.Duration,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		tokenGen:        tokenGen,
		hashProvider:    hashProvider,
		refreshRepo:     refreshRepo,
		sessionDuration: sessionDuration,
	}
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		masked := maskEmail(email)
		slog.ErrorContext(ctx, "falha ao buscar usuario por email", slog.String("email", masked), slog.String("error", err.Error()))
		return "", domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		masked := maskEmail(user.Email)
		slog.WarnContext(ctx, "tentativa de login em conta inativa", slog.String("email", masked), slog.String("user_id", user.ID.String()))
		return "", domain.ErrInvalidCredentials
	}

	err = uc.hashProvider.CompareHash(user.PasswordHash, password)
	if err != nil {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "senha incorreta", slog.String("email", masked), slog.String("user_id", user.ID.String()))
		return "", domain.ErrInvalidCredentials
	}

	token, err := uc.tokenGen.GenerateToken(ctx, user)
	if err != nil {
		masked := maskEmail(user.Email)
		slog.ErrorContext(ctx, "falha ao gerar token", slog.String("email", masked), slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("erro ao gerar token: %w", err)
	}

	err = uc.refreshRepo.Save(ctx, user.ID, token, uc.sessionDuration)
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

func (uc *authUseCase) Logout(ctx context.Context, token string) error {
	err := uc.refreshRepo.Revoke(ctx, token)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao encerrar sessão", slog.String("error", err.Error()))
		return fmt.Errorf("erro ao revogar refresh token: %w", err)
	}
	slog.InfoContext(ctx, "sessão encerrada com sucesso")
	return nil
}

func (uc *authUseCase) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	err := uc.refreshRepo.RevokeAll(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao revogar todas as sessões", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		return fmt.Errorf("erro ao revogar todas as sessões: %w", err)
	}
	slog.InfoContext(ctx, "todas as sessões revogadas com sucesso", slog.String("user_id", userID.String()))
	return nil
}

func (uc *authUseCase) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	sessions, err := uc.refreshRepo.ListByUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao listar sessões", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao listar sessões: %w", err)
	}
	return sessions, nil
}

func (uc *authUseCase) RevokeSession(ctx context.Context, sessionID string) error {
	err := uc.refreshRepo.RevokeByID(ctx, sessionID)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao revogar sessão", slog.String("session_id", sessionID), slog.String("error", err.Error()))
		return fmt.Errorf("erro ao revogar sessão: %w", err)
	}
	slog.InfoContext(ctx, "sessão revogada com sucesso", slog.String("session_id", sessionID))
	return nil
}
