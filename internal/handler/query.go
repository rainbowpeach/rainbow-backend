package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

// QueryHandler serves scene queries whose scene_code is supplied explicitly by
// the client. These handlers intentionally do not resolve the request Host.
type QueryHandler struct {
	contentService         *service.ContentService
	scenePageConfigService *service.ScenePageConfigService
}

func NewQueryHandler(contentService *service.ContentService, scenePageConfigService *service.ScenePageConfigService) *QueryHandler {
	return &QueryHandler{
		contentService:         contentService,
		scenePageConfigService: scenePageConfigService,
	}
}

func (h *QueryHandler) GetScenePageConfig(c *gin.Context) {
	sceneCode := strings.TrimSpace(c.Query("scene_code"))
	if sceneCode == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.scenePageConfigService.GetBySceneCode(c.Request.Context(), sceneCode)
	if err != nil {
		h.respondScenePageConfigError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *QueryHandler) GetContent(c *gin.Context) {
	sceneCode := strings.TrimSpace(c.Query("scene_code"))
	date := strings.TrimSpace(c.Query("date"))
	if sceneCode == "" || date == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.contentService.GetBySceneAndDate(c.Request.Context(), sceneCode, date)
	if err != nil {
		h.respondContentError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *QueryHandler) respondScenePageConfigError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidScenePageConfigParams):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrScenePageConfigNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeScenePageConfigNotFound, "scene page config not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}

func (h *QueryHandler) respondContentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidContentParams):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrInvalidDateFormat):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidDateFormat, "invalid date format")
	case errors.Is(err, service.ErrContentNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeContentNotFound, "content not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}
