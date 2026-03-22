package handler

import (
	"net/http"
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

type roleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toRoleResponse(role *domain.Role) roleResponse {
	return roleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
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

	role, err := h.roleUseCase.FindRoleByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	return c.JSON(http.StatusOK, toRoleResponse(role))
}

// List (GET /roles)
func (h *RoleHandler) List(c *echo.Context) error {
	roles, err := h.roleUseCase.ListRoles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list roles")
	}

	resp := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, toRoleResponse(role))
	}
	return c.JSON(http.StatusOK, resp)
}

// PartialUpdate (PATCH /roles/:id)
func (h *RoleHandler) PartialUpdate(c *echo.Context) error {
	idStr := c.Param("id") // CORRIGIDO
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	var req updateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}

	ctx := c.Request().Context()

	role, err := h.roleUseCase.FindRoleByID(ctx, id)
	if err != nil {
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

// Delete (DELETE /roles/:id)
func (h *RoleHandler) Delete(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	if err := h.roleUseCase.DeleteRole(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao excluir perfil")
	}

	return c.NoContent(http.StatusNoContent)
}
