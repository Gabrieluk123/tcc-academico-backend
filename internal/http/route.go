package http

import (
	"academico/internal/http/handler"

	"github.com/labstack/echo/v5"
)

func SetupRoutes(
	e *echo.Echo,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
) {
	api := e.Group("/api/v1")

	api.POST("/login", authHandler.Login)

	api.POST("/users", userHandler.Create)
	api.GET("/users", userHandler.List)
	api.GET("/users/:id", userHandler.GetByID)
	api.PATCH("/users/:id", userHandler.PartialUpdate)
	api.DELETE("/users/:id", userHandler.Deactivate)

	api.POST("/roles", roleHandler.Create)
	api.GET("/roles", roleHandler.List)
	api.GET("/roles/:id", roleHandler.GetByID)
	api.PATCH("/roles/:id", roleHandler.PartialUpdate)
	api.DELETE("/roles/:id", roleHandler.Delete)
}
