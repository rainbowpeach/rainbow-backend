package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type AdminAuthHandler struct {
	authService *service.AuthService
}

func NewAdminAuthHandler(authService *service.AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.CodeInternalServerError, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}
