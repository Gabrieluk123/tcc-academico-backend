package auth

import (
	"context"
	"fmt"
	"log/slog"

	"academico/internal/domain"

	"github.com/google/uuid"
)

var _ domain.PermissionUseCase = (*permissionUseCase)(nil)

type permissionUseCase struct {
	repo domain.PermissionRepository
}

func NewPermissionUseCase(repo domain.PermissionRepository) domain.PermissionUseCase {
	return &permissionUseCase{repo: repo}
}

func (uc *permissionUseCase) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	permissions, err := uc.repo.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao listar permissões", slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao listar permissões: %w", err)
	}
	return permissions, nil
}

func (uc *permissionUseCase) FindPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	permission, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar permissão", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao buscar permissão: %w", err)
	}
	return permission, nil
}
