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
		return fmt.Errorf("erro ao hashear senha: %w", err)
	}
	user.PasswordHash = hash
	return u.repo.Create(ctx, user)
}

func (u *userUseCase) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *userUseCase) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.repo.FindByEmail(ctx, email)
}

func (u *userUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	return u.repo.Update(ctx, user)
}

func (u *userUseCase) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return u.repo.List(ctx)
}

func (u *userUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	user.IsActive = false
	now := time.Now()
	user.DisabledAt = &now
	slog.InfoContext(ctx, "usuário inativado", slog.String("user_id", user.ID.String()))
	return u.repo.Update(ctx, user)
}
