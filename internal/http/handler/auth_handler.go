package handler

import (
	"errors"
	"net/http"
	"time"

	"academico/internal/domain"

	"github.com/google/uuid"
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

func (h *AuthHandler) Logout(c *echo.Context) error {
	token, ok := c.Get("refresh_token").(string)
	if !ok || token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh token ausente")
	}

	ctx := c.Request().Context()
	if err := h.useCase.Logout(ctx, token); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao revogar sessão: "+err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) ListSessions(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id inválido")
	}
	ctx := c.Request().Context()
	sessions, err := h.useCase.ListSessions(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao listar sessões")
	}
	type sessionResponse struct {
		ID        string    `json:"id"`
		UserID    string    `json:"user_id"`
		ExpiresAt time.Time `json:"expires_at"`
		Revoked   bool      `json:"revoked"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	resp := make([]sessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, sessionResponse{
			ID:        s.ID,
			UserID:    s.UserID.String(),
			ExpiresAt: s.ExpiresAt,
			Revoked:   s.Revoked,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"items": resp, "total": len(resp)})
}

// LogoutAll (DELETE /users/:id/sessions)
func (h *AuthHandler) LogoutAll(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id inválido")
	}
	ctx := c.Request().Context()
	if err := h.useCase.LogoutAll(ctx, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao revogar sessões")
	}
	return c.NoContent(http.StatusNoContent)
}

// RevokeSession (DELETE /sessions/:id)
func (h *AuthHandler) RevokeSession(c *echo.Context) error {
	sessionID := c.Param("id")
	ctx := c.Request().Context()
	if err := h.useCase.RevokeSession(ctx, sessionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao revogar sessão")
	}
	return c.NoContent(http.StatusNoContent)
}
