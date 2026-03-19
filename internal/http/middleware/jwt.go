package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyRoleID = "role_id"
)

// NewJWTMiddleware returns an Echo middleware that validates a Bearer JWT and
// stores user_id and role_id claims in the echo.Context for downstream handlers.
func NewJWTMiddleware(secret string) echo.MiddlewareFunc {
	keyBytes := []byte(secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "token de autenticação ausente")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "algoritmo de assinatura inválido")
				}
				return keyBytes, nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "token inválido ou expirado")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "claims inválidos")
			}

			userID, _ := claims[ContextKeyUserID].(string)
			roleID, _ := claims[ContextKeyRoleID].(string)

			if userID == "" || roleID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "token sem identificadores de usuário")
			}

			c.Set(ContextKeyUserID, userID)
			c.Set(ContextKeyRoleID, roleID)

			return next(c)
		}
	}
}
