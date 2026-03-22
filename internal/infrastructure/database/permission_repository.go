package database

import (
	"context"
	"fmt"

	"academico/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

var _ domain.PermissionRepository = (*permissionRepository)(nil)

func NewPermissionRepository(db *gorm.DB) *permissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	dbPerm := FromDomainPermission(permission)
	if err := r.db.WithContext(ctx).Create(dbPerm).Error; err != nil {
		return fmt.Errorf("erro ao criar permissão: %w", err)
	}
	return nil
}

func (r *permissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var dbPerm PermissionDB
	err := r.db.WithContext(ctx).First(&dbPerm, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar permissão por ID: %w", err)
	}
	return dbPerm.ToDomain(), nil
}

func (r *permissionRepository) FindBySlug(ctx context.Context, slug string) (*domain.Permission, error) {
	var dbPerm PermissionDB
	err := r.db.WithContext(ctx).First(&dbPerm, "slug = ?", slug).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar permissão: %w", err)
	}
	return dbPerm.ToDomain(), nil
}

func (r *permissionRepository) List(ctx context.Context) ([]*domain.Permission, error) {
	var dbPerms []PermissionDB
	if err := r.db.WithContext(ctx).Order("slug").Find(&dbPerms).Error; err != nil {
		return nil, fmt.Errorf("erro ao listar permissões: %w", err)
	}
	perms := make([]*domain.Permission, 0, len(dbPerms))
	for _, p := range dbPerms {
		perms = append(perms, p.ToDomain())
	}
	return perms, nil
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&PermissionDB{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("erro ao excluir permissão: %w", err)
	}
	return nil
}
