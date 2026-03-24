package handler

import (
	"fmt"
	"net/http"
	"strings"

	"academico/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type PermissionHandler struct {
	useCase domain.PermissionUseCase
}

func NewPermissionHandler(useCase domain.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{useCase: useCase}
}

type permissionResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// List (GET /permissions)
func (h *PermissionHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	params := domain.PermissionListParams{}

	if page := c.QueryParam("page"); page != "" {
		fmt.Sscanf(page, "%d", &params.Page)
	}
	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		fmt.Sscanf(pageSize, "%d", &params.PageSize)
	}
	params.OrderBy = c.QueryParam("order_by")
	params.OrderDir = c.QueryParam("order_dir")
	if slug := c.QueryParam("slug"); slug != "" {
		s := strings.TrimSpace(slug)
		params.Slug = &s
	}

	permissions, total, err := h.useCase.ListPermissions(ctx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list permissions")
	}

	resp := make([]permissionResponse, 0, len(permissions))
	for _, p := range permissions {
		resp = append(resp, permissionResponse{
			ID:          p.ID.String(),
			Slug:        p.Slug,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// GetByID (GET /permissions/:id)
func (h *PermissionHandler) GetByID(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	ctx := c.Request().Context()
	permission, err := h.useCase.FindPermissionByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find permission")
	}
	if permission == nil {
		return echo.NewHTTPError(http.StatusNotFound, "permission not found")
	}

	return c.JSON(http.StatusOK, permissionResponse{
		ID:          permission.ID.String(),
		Slug:        permission.Slug,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   permission.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
