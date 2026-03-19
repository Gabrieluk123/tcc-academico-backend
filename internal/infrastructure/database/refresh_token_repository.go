package database

import (
	"context"
	"time"

	"academico/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

var _ domain.RefreshTokenRepository = (*refreshTokenRepository)(nil)

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(ctx context.Context, userID uuid.UUID, token string, expiresIn time.Duration) error {
	dbToken := &RefreshTokenDB{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		Revoked:   false,
	}
	return r.db.WithContext(ctx).Create(dbToken).Error
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&RefreshTokenDB{}).
		Where("token = ?", token).
		Update("revoked", true).
		Error
}
