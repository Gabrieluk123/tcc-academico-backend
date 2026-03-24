package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"academico/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type RoleHandler struct {
	roleUseCase domain.RoleUseCase
}

func NewRoleHandler(roleUseCase domain.RoleUseCase) *RoleHandler {
	return &RoleHandler{roleUseCase: roleUseCase}
}

// ================= DTOs =================
type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateRoleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type permissionInRoleResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type roleResponse struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Permissions []*permissionInRoleResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

func toRoleResponse(role *domain.Role) roleResponse {
	resp := roleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
	if len(role.Permissions) > 0 {
		resp.Permissions = make([]*permissionInRoleResponse, 0, len(role.Permissions))
		for _, p := range role.Permissions {
			resp.Permissions = append(resp.Permissions, &permissionInRoleResponse{
				ID:          p.ID.String(),
				Slug:        p.Slug,
				Description: p.Description,
			})
		}
	}
	return resp
}

// parseRoleIncludes splits a comma-separated "include" query param into a lookup set for roles.
func parseRoleIncludes(c *echo.Context) map[string]bool {
	set := make(map[string]bool)
	for _, raw := range c.Request().URL.Query()["include"] {
		for _, item := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(item); v != "" {
				set[v] = true
			}
		}
	}
	return set
}

// ================= HANDLERS =================

// Create (POST /roles)
func (h *RoleHandler) Create(c *echo.Context) error {
	var req createRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	role := &domain.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.roleUseCase.CreateRole(c.Request().Context(), role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create role")
	}

	return c.JSON(http.StatusCreated, toRoleResponse(role))
}

// GetByID (GET /roles/:id)
func (h *RoleHandler) GetByID(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	includes := parseRoleIncludes(c)
	role, err := h.roleUseCase.FindRoleByID(c.Request().Context(), id, includes["permissions"])
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find role")
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	return c.JSON(http.StatusOK, toRoleResponse(role))
}

// List (GET /roles)
func (h *RoleHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	params := domain.RoleListParams{}

	if page := c.QueryParam("page"); page != "" {
		fmt.Sscanf(page, "%d", &params.Page)
	}
	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		fmt.Sscanf(pageSize, "%d", &params.PageSize)
	}
	params.OrderBy = c.QueryParam("order_by")
	params.OrderDir = c.QueryParam("order_dir")
	if name := c.QueryParam("name"); name != "" {
		params.Name = &name
	}

	includes := parseRoleIncludes(c)
	params.IncludePermissions = includes["permissions"]

	roles, total, err := h.roleUseCase.ListRoles(ctx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list roles")
	}

	resp := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, toRoleResponse(role))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// PartialUpdate (PATCH /roles/:id)
func (h *RoleHandler) PartialUpdate(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	var req updateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}

	ctx := c.Request().Context()

	role, err := h.roleUseCase.FindRoleByID(ctx, id, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao buscar perfil")
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "perfil não encontrado")
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := h.roleUseCase.UpdateRole(ctx, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao atualizar perfil")
	}

	return c.JSON(http.StatusOK, toRoleResponse(role))
}

// ReassignUsers (PATCH /roles/:id/users)
func (h *RoleHandler) ReassignUsers(c *echo.Context) error {
	idStr := c.Param("id")
	fromID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	var req struct {
		NewRoleID uuid.UUID `json:"new_role_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}
	if req.NewRoleID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "new_role_id é obrigatório")
	}

	if err := h.roleUseCase.ReassignUsers(c.Request().Context(), fromID, req.NewRoleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao reatribuir usuários")
	}

	return c.NoContent(http.StatusNoContent)
}

// Delete (DELETE /roles/:id)
func (h *RoleHandler) Delete(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	if err := h.roleUseCase.DeleteRole(c.Request().Context(), id); err != nil {
		if err == domain.ErrRoleInUse {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao excluir perfil")
	}

	return c.NoContent(http.StatusNoContent)
}
