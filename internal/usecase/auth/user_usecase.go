package auth

import (
	"context"
	"fmt"
	"time"

	"academico/internal/domain"
	"log/slog"

	"github.com/google/uuid"
)

var _ domain.UserUseCase = (*userUseCase)(nil)

type userUseCase struct {
	repo     domain.UserRepository
	hashProv domain.HashProvider
}

func NewUserUseCase(repo domain.UserRepository, hashProv domain.HashProvider) domain.UserUseCase {
	return &userUseCase{repo: repo, hashProv: hashProv}
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	hash, err := u.hashProv.HashPassword(user.PasswordHash)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao hashear senha do usuário",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao hashear senha: %w", err)
	}
	user.PasswordHash = hash
	if err := u.repo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "falha ao criar usuário",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao criar usuário: %w", err)
	}
	slog.InfoContext(ctx, "usuário criado com sucesso",
		slog.String("user_id", user.ID.String()),
	)
	return nil
}

func (u *userUseCase) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar usuário por ID",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	return user, nil
}

func (u *userUseCase) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar usuário por email",
			slog.String("email", maskEmail(email)),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao buscar usuário por email: %w", err)
	}
	return user, nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	if err := u.repo.Update(ctx, user); err != nil {
		slog.ErrorContext(ctx, "falha ao atualizar usuário",
			slog.String("user_id", user.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao atualizar usuário: %w", err)
	}
	slog.InfoContext(ctx, "usuário atualizado com sucesso",
		slog.String("user_id", user.ID.String()),
	)
	return nil
}

func (u *userUseCase) ListUsers(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	users, total, err := u.repo.List(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao listar usuários",
			slog.String("error", err.Error()),
		)
		return nil, 0, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	return users, total, nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar usuário para inativar",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	user.IsActive = false
	now := time.Now()
	user.DisabledAt = &now
	if err := u.repo.Update(ctx, user); err != nil {
		slog.ErrorContext(ctx, "falha ao inativar usuário",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao inativar usuário: %w", err)
	}
	slog.InfoContext(ctx, "usuário inativado com sucesso",
		slog.String("user_id", user.ID.String()),
	)
	return nil
}
