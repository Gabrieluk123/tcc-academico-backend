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

func (r *permissionRepository) List(ctx context.Context, params domain.PermissionListParams) ([]*domain.Permission, int, error) {
	db := r.db.WithContext(ctx).Model(&PermissionDB{})

	if params.Slug != nil {
		db = db.Where("slug ILIKE ?", "%"+*params.Slug+"%")
	}

	// Total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar permissões: %w", err)
	}

	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100 // permissions are few, default to a larger page
	}
	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{"id": true, "slug": true, "created_at": true}
	orderBy := params.OrderBy
	orderDir := params.OrderDir
	if !allowedOrderBy[orderBy] {
		orderBy = "slug"
	}
	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "asc"
	}

	var dbPerms []PermissionDB
	if err := db.Offset(offset).Limit(pageSize).Order(orderBy + " " + orderDir).Find(&dbPerms).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao listar permissões: %w", err)
	}
	perms := make([]*domain.Permission, 0, len(dbPerms))
	for _, p := range dbPerms {
		perms = append(perms, p.ToDomain())
	}
	return perms, int(total), nil
}

// ListByRoleName resolves the slugs that Casbin has granted to the given role name
// and returns the matching Permission records.
func (r *permissionRepository) ListByRoleName(ctx context.Context, roleName string) ([]*domain.Permission, error) {
	// The casbin_rule table stores: ptype='p', v0=roleName, v1=resource, v2=action
	// We reconstruct slug as "resource:action" and match against the permissions table.
	type casbinRow struct {
		V1 string
		V2 string
	}
	var rows []casbinRow
	if err := r.db.WithContext(ctx).Raw(
		`SELECT v1, v2 FROM casbin_rule WHERE ptype = 'p' AND v0 = ?`, roleName,
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar policies do Casbin: %w", err)
	}

	if len(rows) == 0 {
		return []*domain.Permission{}, nil
	}

	slugs := make([]string, 0, len(rows))
	hasWildcard := false
	for _, row := range rows {
		if row.V1 == "*" || row.V2 == "*" {
			hasWildcard = true
			continue
		}
		slugs = append(slugs, row.V1+":"+row.V2)
	}

	// A wildcard policy (e.g. Admin:*:*) means full access — return all permissions.
	if hasWildcard {
		var dbPerms []PermissionDB
		if err := r.db.WithContext(ctx).Find(&dbPerms).Error; err != nil {
			return nil, fmt.Errorf("erro ao buscar todas as permissões: %w", err)
		}
		perms := make([]*domain.Permission, 0, len(dbPerms))
		for _, p := range dbPerms {
			perms = append(perms, p.ToDomain())
		}
		return perms, nil
	}

	if len(slugs) == 0 {
		return []*domain.Permission{}, nil
	}

	var dbPerms []PermissionDB
	if err := r.db.WithContext(ctx).Where("slug IN ?", slugs).Find(&dbPerms).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar permissões por slug: %w", err)
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
