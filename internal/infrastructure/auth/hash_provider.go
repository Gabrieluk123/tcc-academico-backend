package auth

import (
	"academico/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var _ domain.HashProvider = (*hashProvider)(nil)

type hashProvider struct{}

func NewHashProvider() domain.HashProvider {
	return &hashProvider{}
}

func (h *hashProvider) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *hashProvider) CompareHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
