// internal/infrastructure/database/role_repository.go
package database

import (
	"context"
	"fmt"

	"academico/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

var _ domain.RoleRepository = (*roleRepository)(nil)

func NewRoleRepository(db *gorm.DB) *roleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	dbRole := FromDomainRole(role)
	if err := r.db.WithContext(ctx).Create(dbRole).Error; err != nil {
		return fmt.Errorf("erro ao criar perfil: %w", err)
	}
	return nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var dbRole RoleDB
	err := r.db.WithContext(ctx).First(&dbRole, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar perfil: %w", err)
	}
	return dbRole.ToDomain(), nil
}

func (r *roleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	var dbRoles []RoleDB
	if err := r.db.WithContext(ctx).Find(&dbRoles).Error; err != nil {
		return nil, fmt.Errorf("erro ao listar perfis: %w", err)
	}
	roles := make([]*domain.Role, 0, len(dbRoles))
	for _, dbRole := range dbRoles {
		roles = append(roles, dbRole.ToDomain())
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	dbRole := FromDomainRole(role)
	if err := r.db.WithContext(ctx).Model(&RoleDB{}).Where("id = ?", role.ID).Updates(dbRole).Error; err != nil {
		return fmt.Errorf("erro ao atualizar perfil: %w", err)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&RoleDB{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("erro ao excluir perfil: %w", err)
	}
	return nil
}
