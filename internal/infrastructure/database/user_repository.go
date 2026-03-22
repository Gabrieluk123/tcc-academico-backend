// internal/infrastructure/database/user_repository.go
package database

import (
	"context"
	"fmt"
	"time"

	"academico/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

var _ domain.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	dbUser := FromDomainUser(user)
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return fmt.Errorf("erro ao criar usuário: %w", err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var dbUser UserDB
	err := r.db.WithContext(ctx).Preload("Role").First(&dbUser, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	return dbUser.ToDomain(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dbUser UserDB
	err := r.db.WithContext(ctx).First(&dbUser, "email = ?", email).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	return dbUser.ToDomain(), nil
}

func (r *userRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int, error) {
	var dbUsers []UserDB
	db := r.db.WithContext(ctx).Model(&UserDB{})

	if params.IncludeRole {
		db = db.Preload("Role")
	}

	// Filtros
	if params.RoleID != nil {
		db = db.Where("role_id = ?", *params.RoleID)
	}
	if params.IsActive != nil {
		db = db.Where("is_active = ?", *params.IsActive)
	}
	if params.FirstName != nil {
		db = db.Where("first_name ILIKE ?", "%"+*params.FirstName+"%")
	}
	if params.LastName != nil {
		db = db.Where("last_name ILIKE ?", "%"+*params.LastName+"%")
	}
	if params.Email != nil {
		db = db.Where("email ILIKE ?", "%"+*params.Email+"%")
	}
	if params.CreatedAt != nil {
		db = db.Where("created_at >= ?", *params.CreatedAt)
	}
	if params.UpdatedAt != nil {
		db = db.Where("updated_at >= ?", *params.UpdatedAt)
	}
	if params.DisabledAt != nil {
		db = db.Where("disabled_at >= ?", *params.DisabledAt)
	}

	// Paginação
	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	// Ordenação — whitelist para prevenir SQL injection
	allowedOrderBy := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"role_id":    true,
	}
	orderBy := params.OrderBy
	orderDir := params.OrderDir
	if !allowedOrderBy[orderBy] {
		orderBy = "id"
	}
	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "desc"
	}

	// Total — conta com os filtros ativos, sem Offset/Limit
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar usuários: %w", err)
	}

	// Busca com paginação e ordenação
	if err := db.Offset(offset).Limit(pageSize).Order(orderBy + " " + orderDir).Find(&dbUsers).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	users := make([]*domain.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.ToDomain())
	}
	return users, int(total), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	dbUser := FromDomainUser(user)
	if err := r.db.WithContext(ctx).Model(&UserDB{}).Where("id = ?", user.ID).Updates(dbUser).Error; err != nil {
		return fmt.Errorf("erro ao atualizar usuário: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&UserDB{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":   false,
			"disabled_at": &now,
		}).Error; err != nil {
		return fmt.Errorf("erro ao desativar usuário: %w", err)
	}
	return nil
}
