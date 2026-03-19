package handler

import (
	"errors"
	"net/http"

	"academico/internal/domain"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	useCase domain.AuthUseCase
}

func NewAuthHandler(useCase domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

func (h *AuthHandler) Login(c *echo.Context) error {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req loginRequest
	if err := c.Bind(&req); err != nil || req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}

	ctx := c.Request().Context()
	token, err := h.useCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "credenciais inválidas")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "erro interno")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}
