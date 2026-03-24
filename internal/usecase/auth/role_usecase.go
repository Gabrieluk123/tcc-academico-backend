package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"academico/internal/domain"

	"github.com/google/uuid"
)

var _ domain.RoleUseCase = (*roleUseCase)(nil)

type roleUseCase struct {
	repo     domain.RoleRepository
	permRepo domain.PermissionRepository
	userRepo domain.UserRepository
	enforcer domain.Enforcer
}

func NewRoleUseCase(repo domain.RoleRepository, permRepo domain.PermissionRepository, userRepo domain.UserRepository, enforcer domain.Enforcer) domain.RoleUseCase {
	return &roleUseCase{repo: repo, permRepo: permRepo, userRepo: userRepo, enforcer: enforcer}
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

func (r *roleUseCase) FindRoleByID(ctx context.Context, id uuid.UUID, includePermissions bool) (*domain.Role, error) {
	role, err := r.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao buscar perfil por ID",
			slog.String("role_id", id.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao buscar perfil: %w", err)
	}
	if role == nil {
		return nil, nil
	}
	if includePermissions {
		perms, err := r.permRepo.ListByRoleName(ctx, role.Name)
		if err != nil {
			slog.ErrorContext(ctx, "falha ao buscar permissões do perfil",
				slog.String("role_id", id.String()),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("erro ao buscar permissões do perfil: %w", err)
		}
		role.Permissions = perms
	}
	return role, nil
}

func (r *roleUseCase) ListRoles(ctx context.Context, params domain.RoleListParams) ([]*domain.Role, int, error) {
	roles, total, err := r.repo.List(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao listar perfis",
			slog.String("error", err.Error()),
		)
		return nil, 0, fmt.Errorf("erro ao listar perfis: %w", err)
	}
	return roles, total, nil
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

func (r *roleUseCase) ReassignUsers(ctx context.Context, fromRoleID, toRoleID uuid.UUID) error {
	count, err := r.userRepo.CountByRoleID(ctx, fromRoleID)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao contar usuários com o perfil",
			slog.String("role_id", fromRoleID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao verificar uso do perfil: %w", err)
	}
	if count == 0 {
		return nil
	}
	if err := r.userRepo.BulkUpdateRoleID(ctx, fromRoleID, toRoleID); err != nil {
		slog.ErrorContext(ctx, "falha ao reatribuir perfil dos usuários",
			slog.String("from_role_id", fromRoleID.String()),
			slog.String("to_role_id", toRoleID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao reatribuir usuários: %w", err)
	}
	slog.InfoContext(ctx, "usuários reatribuídos ao novo perfil",
		slog.String("from_role_id", fromRoleID.String()),
		slog.String("to_role_id", toRoleID.String()),
		slog.Int("users_count", count),
	)
	return nil
}

func (r *roleUseCase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	count, err := r.userRepo.CountByRoleID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "falha ao contar usuários com o perfil",
			slog.String("role_id", id.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao verificar uso do perfil: %w", err)
	}
	if count > 0 {
		slog.WarnContext(ctx, "tentativa de excluir perfil ainda em uso",
			slog.String("role_id", id.String()),
			slog.Int("users_count", count),
		)
		return domain.ErrRoleInUse
	}
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

func (r *roleUseCase) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	role, err := r.repo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("erro ao buscar perfil: %w", err)
	}
	if role == nil {
		return nil
	}

	if err := r.enforcer.RemoveAllPoliciesForRole(ctx, role.Name); err != nil {
		slog.ErrorContext(ctx, "falha ao remover policies do perfil",
			slog.String("role_id", roleID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("erro ao limpar permissões do perfil: %w", err)
	}

	for _, permID := range permissionIDs {
		perm, err := r.permRepo.FindByID(ctx, permID)
		if err != nil {
			return fmt.Errorf("erro ao buscar permissão %s: %w", permID, err)
		}
		if perm == nil {
			return fmt.Errorf("permissão %s não encontrada", permID)
		}
		parts := strings.SplitN(perm.Slug, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("slug de permissão inválido: %s", perm.Slug)
		}
		if _, err := r.enforcer.AddPolicy(ctx, role.Name, parts[0], parts[1]); err != nil {
			slog.ErrorContext(ctx, "falha ao adicionar policy",
				slog.String("role_id", roleID.String()),
				slog.String("slug", perm.Slug),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("erro ao atribuir permissão %s: %w", perm.Slug, err)
		}
	}

	slog.InfoContext(ctx, "permissões do perfil atualizadas",
		slog.String("role_id", roleID.String()),
		slog.Int("permissions_count", len(permissionIDs)),
	)
	return nil
}
