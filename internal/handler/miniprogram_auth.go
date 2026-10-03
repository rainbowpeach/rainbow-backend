package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type MiniProgramAuthHandler struct {
	authService *service.MiniProgramAuthService
}

func NewMiniProgramAuthHandler(authService *service.MiniProgramAuthService) *MiniProgramAuthHandler {
	return &MiniProgramAuthHandler{authService: authService}
}

func (h *MiniProgramAuthHandler) Login(c *gin.Context) {
	var req model.MiniProgramLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("mini program login request invalid ip=%s err=%v", c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMiniProgramCode):
			log.Printf("mini program login code invalid ip=%s", c.ClientIP())
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		case errors.Is(err, service.ErrWeChatLoginFailed):
			log.Printf("mini program login wechat verification failed ip=%s err=%v", c.ClientIP(), err)
			model.WriteError(c, http.StatusUnauthorized, model.CodeUnauthorized, "unauthorized")
		default:
			log.Printf("mini program login internal error ip=%s err=%v", c.ClientIP(), err)
			model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
		}
		return
	}

	log.Printf("mini program login succeeded ip=%s", c.ClientIP())
	model.WriteOK(c, result)
}
