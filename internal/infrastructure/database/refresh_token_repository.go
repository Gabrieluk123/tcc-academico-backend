package database

import (
	"context"
	"errors"
	"fmt"
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
	if err := r.db.WithContext(ctx).Create(dbToken).Error; err != nil {
		return fmt.Errorf("erro ao salvar refresh token para user_id %s: %w", userID.String(), err)
	}
	return nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenDB{}).
		Where("token = ?", token).
		Update("revoked", true).
		Error; err != nil {
		return fmt.Errorf("erro ao revogar refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenDB{}).
		Where("user_id = ? AND revoked = false", userID.String()).
		Update("revoked", true).
		Error; err != nil {
		return fmt.Errorf("erro ao revogar todas as sessões do user_id %s: %w", userID.String(), err)
	}
	return nil
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	var dbToken RefreshTokenDB
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&dbToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("erro ao buscar sessão pelo token: %w", err)
	}
	userID, err := uuid.Parse(dbToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user_id inválido na sessão: %w", err)
	}
	return &domain.Session{
		ID:        dbToken.ID,
		UserID:    userID,
		Token:     dbToken.Token,
		ExpiresAt: dbToken.ExpiresAt,
		Revoked:   dbToken.Revoked,
		CreatedAt: dbToken.CreatedAt,
		UpdatedAt: dbToken.UpdatedAt,
	}, nil
}

func (r *refreshTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	var tokens []RefreshTokenDB
	err := r.db.WithContext(ctx).
		Model(&RefreshTokenDB{}).
		Where("user_id = ?", userID.String()).
		Find(&tokens).Error
	if err != nil {
		return nil, fmt.Errorf("erro ao listar sessões do user_id %s: %w", userID.String(), err)
	}
	sessions := make([]domain.Session, 0, len(tokens))
	for _, t := range tokens {
		sessions = append(sessions, domain.Session{
			ID:        t.ID,
			UserID:    userID,
			Token:     t.Token,
			ExpiresAt: t.ExpiresAt,
			Revoked:   t.Revoked,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}
	return sessions, nil
}

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, sessionID string) error {
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenDB{}).
		Where("id = ?", sessionID).
		Update("revoked", true).
		Error; err != nil {
		return fmt.Errorf("erro ao revogar sessão %s: %w", sessionID, err)
	}
	return nil
}
