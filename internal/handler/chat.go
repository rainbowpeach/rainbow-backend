package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type ChatHandler struct {
	chat *service.ChatService
}

func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("user chat invalid request user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.chat.Reply(c.Request.Context(), userID, &req)
	if err != nil {
		log.Printf("user chat failed user_id=%d ip=%s scene=%s mode=%s err=%v", userID, c.ClientIP(), req.SceneCode, req.Mode, err)
		h.respondChatError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *ChatHandler) ListLogs(c *gin.Context) {
	var req model.ChatLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("admin chat log list invalid query %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.chat.ListLogs(c.Request.Context(), req)
	if err != nil {
		log.Printf("admin chat log list failed %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		h.respondChatError(c, err)
		return
	}

	log.Printf("admin chat logs listed %s ip=%s total=%d", adminActor(c), c.ClientIP(), result.Total)
	model.WriteOK(c, result)
}

func (h *ChatHandler) respondChatError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidChat), errors.Is(err, service.ErrChatPersona):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrChatRateLimited):
		model.WriteError(c, http.StatusTooManyRequests, model.CodeChatRateLimited, "too many requests")
	case errors.Is(err, service.ErrChatNotConfigured), errors.Is(err, service.ErrChatUpstream):
		model.WriteError(c, http.StatusServiceUnavailable, model.CodeChatUnavailable, "chat unavailable")
	case errors.Is(err, service.ErrUserNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeUserNotFound, "user not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}
