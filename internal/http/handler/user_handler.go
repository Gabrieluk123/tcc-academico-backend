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

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(userUseCase domain.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// ================= DTOs =================
type createUserRequest struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	RoleID    uuid.UUID `json:"role_id"`
}

type updateUserRequest struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Email     *string    `json:"email"`
	RoleID    *uuid.UUID `json:"role_id"`
	IsActive  *bool      `json:"is_active"`
}

type userResponse struct {
	ID        string              `json:"id"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	Email     string              `json:"email"`
	RoleID    string              `json:"role_id"`
	Role      *roleInUserResponse `json:"role,omitempty"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type roleInUserResponse struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Permissions []*permissionInRoleResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

func toUserResponse(user *domain.User, includeRole bool, includePermissions bool) userResponse {
	resp := userResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		RoleID:    user.RoleID.String(),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if includeRole && user.Role != nil {
		roleResp := &roleInUserResponse{
			ID:          user.Role.ID.String(),
			Name:        user.Role.Name,
			Description: user.Role.Description,
			CreatedAt:   user.Role.CreatedAt,
			UpdatedAt:   user.Role.UpdatedAt,
		}
		if includePermissions && len(user.Role.Permissions) > 0 {
			roleResp.Permissions = make([]*permissionInRoleResponse, 0, len(user.Role.Permissions))
			for _, p := range user.Role.Permissions {
				roleResp.Permissions = append(roleResp.Permissions, &permissionInRoleResponse{
					ID:          p.ID.String(),
					Slug:        p.Slug,
					Description: p.Description,
				})
			}
		}
		resp.Role = roleResp
	}
	return resp
}

// parseIncludes splits a comma-separated "include" query param into a lookup set.
// Supports: ?include=role  ?include=role,sessions  ?include=role&include=sessions
func parseIncludes(c *echo.Context) map[string]bool {
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

// Create (POST /users)
func (h *UserHandler) Create(c *echo.Context) error {
	var req createUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	user := &domain.User{
		ID:           uuid.New(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: req.Password,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := h.userUseCase.CreateUser(c.Request().Context(), user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, toUserResponse(user, false, false))
}

// GetByID (GET /users/:id)
func (h *UserHandler) GetByID(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	includes := parseIncludes(c)
	user, err := h.userUseCase.FindUserByID(c.Request().Context(), id, includes["permissions"])
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, toUserResponse(user, includes["role"], includes["permissions"]))
}

// List (GET /users)
func (h *UserHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	params := domain.UserListParams{}

	// Paginação
	if page := c.QueryParam("page"); page != "" {
		fmt.Sscanf(page, "%d", &params.Page)
	}
	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		fmt.Sscanf(pageSize, "%d", &params.PageSize)
	}
	// Ordenação
	params.OrderBy = c.QueryParam("order_by")
	params.OrderDir = c.QueryParam("order_dir")
	// Filtros
	if roleID := c.QueryParam("role_id"); roleID != "" {
		id, err := uuid.Parse(roleID)
		if err == nil {
			params.RoleID = &id
		}
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		var b bool
		fmt.Sscanf(isActive, "%t", &b)
		params.IsActive = &b
	}
	if firstName := c.QueryParam("first_name"); firstName != "" {
		params.FirstName = &firstName
	}
	if lastName := c.QueryParam("last_name"); lastName != "" {
		params.LastName = &lastName
	}
	if email := c.QueryParam("email"); email != "" {
		params.Email = &email
	}

	if createdAt := c.QueryParam("created_at"); createdAt != "" {
		t, err := parseTime(createdAt)
		if err == nil {
			params.CreatedAt = &t
		}
	}
	if updatedAt := c.QueryParam("updated_at"); updatedAt != "" {
		t, err := parseTime(updatedAt)
		if err == nil {
			params.UpdatedAt = &t
		}
	}
	if disabledAt := c.QueryParam("disabled_at"); disabledAt != "" {
		t, err := parseTime(disabledAt)
		if err == nil {
			params.DisabledAt = &t
		}
	}
	includes := parseIncludes(c)
	params.IncludeRole = includes["role"]
	params.IncludePermissions = includes["permissions"]

	users, total, err := h.userUseCase.ListUsers(ctx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list users")
	}

	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, toUserResponse(user, params.IncludeRole, params.IncludePermissions))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// parseTime tenta converter string para time.Time em RFC3339
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// PartialUpdate (PATCH /users/:id)
func (h *UserHandler) PartialUpdate(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req updateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	ctx := c.Request().Context()

	user, err := h.userUseCase.FindUserByID(ctx, id, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.userUseCase.UpdateUser(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
	}

	return c.JSON(http.StatusOK, toUserResponse(user, false, false))
}

// Deactivate (DELETE /users/:id)
func (h *UserHandler) Deactivate(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}

	if err := h.userUseCase.DeleteUser(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao inativar usuário")
	}

	return c.NoContent(http.StatusNoContent)
}

// currentUserID extracts the authenticated user's UUID from the Echo context.
func currentUserID(c *echo.Context) (uuid.UUID, error) {
	raw, ok := c.Get("user_id").(string)
	if !ok || raw == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "não autenticado")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "identificador de usuário inválido no token")
	}
	return id, nil
}

// Me (GET /users/me)
func (h *UserHandler) Me(c *echo.Context) error {
	id, err := currentUserID(c)
	if err != nil {
		return err
	}

	includes := parseIncludes(c)
	user, err := h.userUseCase.FindUserByID(c.Request().Context(), id, includes["permissions"])
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao buscar usuário")
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "usuário não encontrado")
	}

	return c.JSON(http.StatusOK, toUserResponse(user, includes["role"], includes["permissions"]))
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword (PATCH /users/me/password)
func (h *UserHandler) ChangePassword(c *echo.Context) error {
	id, err := currentUserID(c)
	if err != nil {
		return err
	}

	var req changePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "dados inválidos")
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "old_password e new_password são obrigatórios")
	}

	if err := h.userUseCase.ChangePassword(c.Request().Context(), id, req.OldPassword, req.NewPassword); err != nil {
		if err == domain.ErrInvalidCredentials {
			return echo.NewHTTPError(http.StatusUnauthorized, "senha atual incorreta")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "erro ao alterar senha")
	}

	return c.NoContent(http.StatusNoContent)
}
