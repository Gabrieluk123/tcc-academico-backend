package auth

import (
	"context"

	"academico/internal/domain"

	"github.com/google/uuid"
)

// Garante em tempo de compilação que roleUseCase implementa domain.RoleUseCase
var _ domain.RoleUseCase = (*roleUseCase)(nil)

type roleUseCase struct {
	repo domain.RoleRepository
}

func NewRoleUseCase(repo domain.RoleRepository) domain.RoleUseCase {
	return &roleUseCase{repo: repo}
}

func (r *roleUseCase) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.repo.Create(ctx, role)
}

func (r *roleUseCase) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return r.repo.FindByID(ctx, id)
}

func (r *roleUseCase) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return r.repo.List(ctx)
}

func (r *roleUseCase) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.repo.Update(ctx, role)
}

func (r *roleUseCase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}
