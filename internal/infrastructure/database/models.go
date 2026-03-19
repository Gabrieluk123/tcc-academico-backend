package database

import (
	"time"

	"academico/internal/domain"

	"github.com/google/uuid"
)

type RefreshTokenDB struct {
	ID        string    `gorm:"primaryKey;type:uuid;column:id"`
	UserID    string    `gorm:"type:uuid;column:user_id;index"`
	Token     string    `gorm:"column:token;uniqueIndex"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	Revoked   bool      `gorm:"column:revoked"`
}

type UserDB struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;column:id"`
	FirstName    string     `gorm:"column:first_name"`
	LastName     string     `gorm:"column:last_name"`
	Email        string     `gorm:"uniqueIndex;column:email"`
	RoleID       uuid.UUID  `gorm:"type:uuid;column:role_id"`
	PasswordHash string     `gorm:"column:password_hash"`
	IsActive     bool       `gorm:"column:is_active"`
	DisabledAt   *time.Time `gorm:"column:disabled_at"`
}

type RoleDB struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;column:id"`
	Name        string    `gorm:"uniqueIndex;column:name"`
	Description string    `gorm:"column:description"`
}

func (u *UserDB) ToDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		RoleID:       u.RoleID,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
		DisabledAt:   u.DisabledAt,
	}
}

func FromDomainUser(u *domain.User) *UserDB {
	return &UserDB{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		RoleID:       u.RoleID,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
		DisabledAt:   u.DisabledAt,
	}
}

func (r *RoleDB) ToDomain() *domain.Role {
	return &domain.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}

func FromDomainRole(r *domain.Role) *RoleDB {
	return &RoleDB{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}
