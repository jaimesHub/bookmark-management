package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/rs/zerolog/log"
)

// UserHandler exposes register endpoint cho user lifecycle.
// Concrete struct (KHÔNG interface) vì handler không cần mock — test
// inject mock UserService trực tiếp qua NewUserHandler.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler — wire qua main.go (T8) với production UserService instance.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register godoc
// @Summary      Register a new user
// @Description  Tạo user account mới với password được bcrypt hash.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest   true  "Register payload"
// @Success      201   {object}  model.RegisterResponse
// @Failure      400   {object}  map[string]string       "invalid request"
// @Failure      409   {object}  map[string]string       "email already registered | username already taken"
// @Failure      500   {object}  map[string]string       "internal server error"
// @Router       /v1/users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "email already registered"})
		case errors.Is(err, service.ErrUsernameAlreadyExists):
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "username already taken"})
		default:
			log.Error().
				Err(err).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Msg("register user failed")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, model.RegisterResponse{
		Data:    *u,
		Message: "Register an user successfully!",
	})
}
