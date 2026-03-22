package middleware

import (
	"context"

	"academico/internal/common"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// NewRequestIDMiddleware gera um UUID único por requisição e o injeta no
// context.Context, permitindo que logs em qualquer camada (use case, repository)
// incluam o request_id automaticamente via slog.ErrorContext / slog.InfoContext.
// O ID também é exposto no header de resposta X-Request-ID.
func NewRequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestID := uuid.New().String()
			ctx := context.WithValue(c.Request().Context(), common.ContextKeyRequestID, requestID)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set("X-Request-ID", requestID)
			return next(c)
		}
	}
}
