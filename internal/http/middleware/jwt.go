package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"academico/internal/common"
	"academico/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyRoleID = "role_id"
)

// NewJWTMiddleware returns an Echo middleware that:
//  1. Validates the Bearer JWT signature and claims.
//  2. Checks the session against the database (not revoked, not expired).
//  3. Verifies the user_id in the DB record matches the JWT claim (defense in depth).
func NewJWTMiddleware(secret string, tokenRepo domain.RefreshTokenRepository) echo.MiddlewareFunc {
	keyBytes := []byte(secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()
			authHeader := c.Request().Header.Get("Authorization")

			if authHeader == "" {
				slog.WarnContext(ctx, "header Authorization ausente")
				return echo.NewHTTPError(http.StatusUnauthorized, "token de autenticação ausente")
			}

			// Aceita "Bearer" em qualquer capitalização (Bearer, bearer, BEARER)
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				slog.WarnContext(ctx, "header Authorization com formato inválido",
					slog.String("scheme_recebido", parts[0]),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "token de autenticação ausente")
			}

			tokenStr := parts[1]

			// 1. Valida assinatura e estrutura do JWT
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "algoritmo de assinatura inválido")
				}
				return keyBytes, nil
			})
			if err != nil || !token.Valid {
				slog.WarnContext(ctx, "token JWT inválido ou expirado", slog.String("error", err.Error()))
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

			// 2. Valida a sessão no banco de dados
			session, err := tokenRepo.FindByToken(ctx, tokenStr)
			if err != nil {
				if errors.Is(err, domain.ErrSessionNotFound) {
					slog.WarnContext(ctx, "tentativa de uso de token sem sessão no banco",
						slog.String("user_id", userID),
					)
					return echo.NewHTTPError(http.StatusUnauthorized, "sessão não encontrada")
				}
				slog.ErrorContext(ctx, "erro ao validar sessão no banco de dados", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "erro interno ao validar sessão")
			}

			if session.Revoked {
				slog.WarnContext(ctx, "tentativa de uso de sessão revogada",
					slog.String("session_id", session.ID),
					slog.String("user_id", userID),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "token inválido ou expirado")
			}

			if time.Now().After(session.ExpiresAt) {
				slog.WarnContext(ctx, "tentativa de uso de sessão expirada",
					slog.String("session_id", session.ID),
					slog.String("user_id", userID),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "token inválido ou expirado")
			}

			// 3. Defesa em profundidade: user_id do banco deve bater com o do token
			if session.UserID.String() != userID {
				slog.WarnContext(ctx, "user_id do token diverge do registrado na sessão",
					slog.String("jwt_user_id", userID),
					slog.String("db_user_id", session.UserID.String()),
					slog.String("session_id", session.ID),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "token inválido ou expirado")
			}

			c.Set(ContextKeyUserID, userID)
			c.Set(ContextKeyRoleID, roleID)

			// Propaga user_id e role_id no context.Context para rastreabilidade nos logs
			reqCtx := context.WithValue(c.Request().Context(), common.ContextKeyUserID, userID)
			reqCtx = context.WithValue(reqCtx, common.ContextKeyRoleID, roleID)
			c.SetRequest(c.Request().WithContext(reqCtx))

			// Disponibiliza o token para uso no logout
			c.Set("refresh_token", tokenStr)

			return next(c)
		}
	}
}
