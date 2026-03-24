package http

import (
	"academico/internal/domain"
	"academico/internal/http/handler"
	"academico/internal/http/middleware"

	"github.com/labstack/echo/v5"
)

func SetupRoutes(
	e *echo.Echo,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	jwtMiddleware echo.MiddlewareFunc,
	enforcer domain.Enforcer,
	roleRepo domain.RoleRepository,
) {
	e.POST("/api/v1/login", authHandler.Login)

	api := e.Group("/api/v1", jwtMiddleware)

	api.POST("/logout", authHandler.Logout)

	api.POST("/users", userHandler.Create,
		middleware.RequirePermission(enforcer, roleRepo, "user", "create"))
	api.GET("/users/me", userHandler.Me)
	api.PATCH("/users/me/password", userHandler.ChangePassword)
	api.GET("/users", userHandler.List,
		middleware.RequirePermission(enforcer, roleRepo, "user", "read"))
	api.GET("/users/:id", userHandler.GetByID,
		middleware.RequirePermission(enforcer, roleRepo, "user", "read"))
	api.PATCH("/users/:id", userHandler.PartialUpdate,
		middleware.RequirePermission(enforcer, roleRepo, "user", "update"))
	api.DELETE("/users/:id", userHandler.Deactivate,
		middleware.RequirePermission(enforcer, roleRepo, "user", "delete"))

	// Sessões de usuários
	api.GET("/users/:id/sessions", authHandler.ListSessions,
		middleware.RequirePermission(enforcer, roleRepo, "session", "read"))
	api.DELETE("/users/:id/sessions", authHandler.LogoutAll,
		middleware.RequirePermission(enforcer, roleRepo, "session", "revoke"))
	api.DELETE("/sessions/:id", authHandler.RevokeSession,
		middleware.RequirePermission(enforcer, roleRepo, "session", "revoke"))

	api.POST("/roles", roleHandler.Create,
		middleware.RequirePermission(enforcer, roleRepo, "role", "create"))
	api.GET("/roles", roleHandler.List,
		middleware.RequirePermission(enforcer, roleRepo, "role", "read"))
	api.GET("/roles/:id", roleHandler.GetByID,
		middleware.RequirePermission(enforcer, roleRepo, "role", "read"))
	api.PATCH("/roles/:id", roleHandler.PartialUpdate,
		middleware.RequirePermission(enforcer, roleRepo, "role", "update"))
	api.PATCH("/roles/:id/users", roleHandler.ReassignUsers,
		middleware.RequirePermission(enforcer, roleRepo, "role", "update"))
	api.DELETE("/roles/:id", roleHandler.Delete,
		middleware.RequirePermission(enforcer, roleRepo, "role", "delete"))

	api.GET("/permissions", permissionHandler.List,
		middleware.RequirePermission(enforcer, roleRepo, "permission", "read"))
	api.GET("/permissions/:id", permissionHandler.GetByID,
		middleware.RequirePermission(enforcer, roleRepo, "permission", "read"))
}
