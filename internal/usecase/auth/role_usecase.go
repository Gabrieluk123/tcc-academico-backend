package auth

import (
	"context"
	"fmt"
	"log/slog"

	"academico/internal/domain"

	"github.com/google/uuid"
)

var _ domain.RoleUseCase = (*roleUseCase)(nil)

type roleUseCase struct {
	repo domain.RoleRepository
}

func NewRoleUseCase(repo domain.RoleRepository) domain.RoleUseCase {
	return &roleUseCase{repo: repo}
}

func (r *roleUseCase) CreateRole(ctx context.Context, role *domain.Role) error {
	if err := r.repo.Create(ctx, role); err != nil {
		slog.ErrorContext(ctx, "falha ao criar perfil",
			slog.String("role_name", role.Name),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao criar perfil: %w", err)
	}
	slog.InfoContext(ctx, "perfil criado com sucesso",
		slog.String("role_id", role.ID.String()),
		slog.String("role_name", role.Name),
	)
	return nil
}

func (r *roleUseCase) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := r.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar perfil por ID",
			slog.String("role_id", id.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao buscar perfil: %w", err)
	}
	return role, nil
}

func (r *roleUseCase) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	roles, err := r.repo.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao listar perfis",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao listar perfis: %w", err)
	}
	return roles, nil
}

func (r *roleUseCase) UpdateRole(ctx context.Context, role *domain.Role) error {
	if err := r.repo.Update(ctx, role); err != nil {
		slog.ErrorContext(ctx, "falha ao atualizar perfil",
			slog.String("role_id", role.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao atualizar perfil: %w", err)
	}
	slog.InfoContext(ctx, "perfil atualizado com sucesso",
		slog.String("role_id", role.ID.String()),
	)
	return nil
}

func (r *roleUseCase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		slog.ErrorContext(ctx, "falha ao excluir perfil",
			slog.String("role_id", id.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao excluir perfil: %w", err)
	}
	slog.InfoContext(ctx, "perfil excluído com sucesso",
		slog.String("role_id", id.String()),
	)
	return nil
}
