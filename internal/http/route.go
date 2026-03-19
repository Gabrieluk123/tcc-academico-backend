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
	jwtMiddleware echo.MiddlewareFunc,
	enforcer domain.Enforcer,
	roleRepo domain.RoleRepository,
) {
	// Public routes — no authentication required.
	e.POST("/api/v1/login", authHandler.Login)

	// Protected group — every route requires a valid JWT.
	api := e.Group("/api/v1", jwtMiddleware)

	// Users
	api.POST("/users", userHandler.Create,
		middleware.RequirePermission(enforcer, roleRepo, "user", "create"))
	api.GET("/users", userHandler.List,
		middleware.RequirePermission(enforcer, roleRepo, "user", "read"))
	api.GET("/users/:id", userHandler.GetByID,
		middleware.RequirePermission(enforcer, roleRepo, "user", "read"))
	api.PATCH("/users/:id", userHandler.PartialUpdate,
		middleware.RequirePermission(enforcer, roleRepo, "user", "update"))
	api.DELETE("/users/:id", userHandler.Deactivate,
		middleware.RequirePermission(enforcer, roleRepo, "user", "delete"))

	// Roles
	api.POST("/roles", roleHandler.Create,
		middleware.RequirePermission(enforcer, roleRepo, "role", "create"))
	api.GET("/roles", roleHandler.List,
		middleware.RequirePermission(enforcer, roleRepo, "role", "read"))
	api.GET("/roles/:id", roleHandler.GetByID,
		middleware.RequirePermission(enforcer, roleRepo, "role", "read"))
	api.PATCH("/roles/:id", roleHandler.PartialUpdate,
		middleware.RequirePermission(enforcer, roleRepo, "role", "update"))
	api.DELETE("/roles/:id", roleHandler.Delete,
		middleware.RequirePermission(enforcer, roleRepo, "role", "delete"))
}
