package http

import (
	"academico/internal/http/handler"

	"github.com/labstack/echo/v5"
)

func SetupRoutes(e *echo.Echo, authHandler *handler.AuthHandler) {
	api := e.Group("/api/v1")

	api.POST("/login", authHandler.Login)
}
