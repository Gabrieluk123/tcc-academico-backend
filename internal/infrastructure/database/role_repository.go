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
	db       *gorm.DB
	permRepo *permissionRepository
}

var _ domain.RoleRepository = (*roleRepository)(nil)

func NewRoleRepository(db *gorm.DB, permRepo *permissionRepository) *roleRepository {
	return &roleRepository{db: db, permRepo: permRepo}
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

func (r *roleRepository) List(ctx context.Context, params domain.RoleListParams) ([]*domain.Role, int, error) {
	db := r.db.WithContext(ctx).Model(&RoleDB{})

	if params.Name != nil {
		db = db.Where("name ILIKE ?", "%"+*params.Name+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar perfis: %w", err)
	}

	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
	orderBy := params.OrderBy
	orderDir := params.OrderDir
	if !allowedOrderBy[orderBy] {
		orderBy = "name"
	}
	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "asc"
	}

	var dbRoles []RoleDB
	if err := db.Offset(offset).Limit(pageSize).Order(orderBy + " " + orderDir).Find(&dbRoles).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao listar perfis: %w", err)
	}

	roles := make([]*domain.Role, 0, len(dbRoles))
	for _, dbRole := range dbRoles {
		role := dbRole.ToDomain()
		if params.IncludePermissions {
			perms, err := r.permRepo.ListByRoleName(ctx, dbRole.Name)
			if err != nil {
				return nil, 0, fmt.Errorf("erro ao buscar permissões do perfil %s: %w", dbRole.Name, err)
			}
			role.Permissions = perms
		}
		roles = append(roles, role)
	}
	return roles, int(total), nil
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
