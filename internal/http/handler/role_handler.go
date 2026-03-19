package handler

import (
	"net/http"

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
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
}

type updateRoleRequest struct {
	Nome      *string `json:"nome"`
	Descricao *string `json:"descricao"`
}

type roleResponse struct {
	ID        string `json:"id"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
}

// ================= HANDLERS =================

// Create (POST /roles)
func (h *RoleHandler) Create(c *echo.Context) error {
	var req createRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}

	role := &domain.Role{
		ID:        uuid.New(),
		Nome:      req.Nome,
		Descricao: req.Descricao,
	}

	if err := h.roleUseCase.CreateRole(c.Request().Context(), role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao criar perfil")
	}

	resp := roleResponse{
		ID:        role.ID.String(),
		Nome:      role.Nome,
		Descricao: role.Descricao,
	}
	return c.JSON(http.StatusCreated, resp)
}

// GetByID (GET /roles/:id)
func (h *RoleHandler) GetByID(c *echo.Context) error {
	idStr := c.Param("id") // CORRIGIDO: Echo v5 padrao string
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	role, err := h.roleUseCase.FindRoleByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "perfil não encontrado")
	}

	resp := roleResponse{
		ID:        role.ID.String(),
		Nome:      role.Nome,
		Descricao: role.Descricao,
	}
	return c.JSON(http.StatusOK, resp)
}

// List (GET /roles)
func (h *RoleHandler) List(c *echo.Context) error {
	roles, err := h.roleUseCase.ListRoles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao listar perfis")
	}

	resp := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, roleResponse{
			ID:        role.ID.String(),
			Nome:      role.Nome,
			Descricao: role.Descricao,
		})
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

	if req.Nome != nil {
		role.Nome = *req.Nome
	}
	if req.Descricao != nil {
		role.Descricao = *req.Descricao
	}

	if err := h.roleUseCase.UpdateRole(ctx, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao atualizar perfil")
	}

	resp := roleResponse{
		ID:        role.ID.String(),
		Nome:      role.Nome,
		Descricao: role.Descricao,
	}
	return c.JSON(http.StatusOK, resp)
}

// Delete (DELETE /roles/:id)
func (h *RoleHandler) Delete(c *echo.Context) error {
	idStr := c.Param("id") // CORRIGIDO
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	if err := h.roleUseCase.DeleteRole(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao excluir perfil")
	}

	return c.NoContent(http.StatusNoContent)
}
