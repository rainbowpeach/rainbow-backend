package handler

import (
	"errors"
	"log"
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
		log.Printf("admin login request invalid ip=%s err=%v", c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			log.Printf("admin login failed username=%q ip=%s", req.Username, c.ClientIP())
			model.WriteError(c, http.StatusUnauthorized, model.CodeUnauthorized, "unauthorized")
		default:
			log.Printf("admin login internal error username=%q ip=%s err=%v", req.Username, c.ClientIP(), err)
			model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
		}
		return
	}

	log.Printf("admin login succeeded username=%q ip=%s", req.Username, c.ClientIP())
	model.WriteOK(c, result)
}
