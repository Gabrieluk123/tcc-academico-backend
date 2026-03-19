package middleware

import (
	"log/slog"
	"net/http"

	"academico/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// RequirePermission returns an Echo middleware that checks whether the
// authenticated user's role has the specified resource+action permission
// via the Casbin enforcer.
//
// The JWT middleware MUST run before this middleware so that role_id is
// already stored in the echo.Context.
func RequirePermission(enforcer domain.Enforcer, roleRepo domain.RoleRepository, resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

			roleIDStr, _ := c.Get(ContextKeyRoleID).(string)
			if roleIDStr == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "role não identificado no token")
			}

			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "role_id inválido no token")
			}

			role, err := roleRepo.FindByID(ctx, roleID)
			if err != nil {
				slog.ErrorContext(ctx, "erro ao buscar role para autorização",
					slog.String("role_id", roleIDStr),
					slog.String("error", err.Error()),
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "erro ao verificar autorização")
			}
			if role == nil {
				return echo.NewHTTPError(http.StatusForbidden, "perfil não encontrado")
			}

			allowed, err := enforcer.Enforce(ctx, role.Name, resource, action)
			if err != nil {
				slog.ErrorContext(ctx, "erro no enforcer Casbin",
					slog.String("role", role.Name),
					slog.String("resource", resource),
					slog.String("action", action),
					slog.String("error", err.Error()),
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "erro ao verificar autorização")
			}

			if !allowed {
				slog.WarnContext(ctx, "acesso negado",
					slog.String("role", role.Name),
					slog.String("resource", resource),
					slog.String("action", action),
				)
				return echo.NewHTTPError(http.StatusForbidden, "acesso negado")
			}

			return next(c)
		}
	}
}
