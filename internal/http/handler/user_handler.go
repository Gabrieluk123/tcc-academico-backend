package handler

import (
	"net/http"

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
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	RoleID    string `json:"role_id"`
	IsActive  bool   `json:"is_active"`
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

	resp := userResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		RoleID:    user.RoleID.String(),
		IsActive:  user.IsActive,
	}
	return c.JSON(http.StatusCreated, resp)
}

// GetByID (GET /users/:id)
func (h *UserHandler) GetByID(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	user, err := h.userUseCase.FindUserByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	resp := userResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		RoleID:    user.RoleID.String(),
		IsActive:  user.IsActive,
	}
	return c.JSON(http.StatusOK, resp)
}

// List (GET /users)
func (h *UserHandler) List(c *echo.Context) error {
	users, err := h.userUseCase.ListUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list users")
	}

	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, userResponse{
			ID:        user.ID.String(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			RoleID:    user.RoleID.String(),
			IsActive:  user.IsActive,
		})
	}
	return c.JSON(http.StatusOK, resp)
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

	user, err := h.userUseCase.FindUserByID(ctx, id)
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

	resp := userResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		RoleID:    user.RoleID.String(),
		IsActive:  user.IsActive,
	}
	return c.JSON(http.StatusOK, resp)
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
