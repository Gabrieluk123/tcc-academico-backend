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
	err := r.db.WithContext(ctx).First(&dbUser, "id = ?", id).Error
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

func (r *userRepository) List(ctx context.Context) ([]*domain.User, error) {
	var dbUsers []UserDB
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&dbUsers).Error; err != nil {
		return nil, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	users := make([]*domain.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.ToDomain())
	}
	return users, nil
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
