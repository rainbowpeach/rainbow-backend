package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/middleware"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type PublicContentHandler struct {
	contentService         *service.ContentService
	scenePageConfigService *service.ScenePageConfigService
	sceneResolver          *service.SceneResolver
	cfg                    config.SceneConfig
}

func NewPublicContentHandler(contentService *service.ContentService, scenePageConfigService *service.ScenePageConfigService, sceneResolver *service.SceneResolver, cfg config.SceneConfig) *PublicContentHandler {
	return &PublicContentHandler{
		contentService:         contentService,
		scenePageConfigService: scenePageConfigService,
		sceneResolver:          sceneResolver,
		cfg:                    cfg,
	}
}

func (h *PublicContentHandler) GetByDate(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	var sceneCode string
	if h.cfg.EnablePublicOverride {
		if override := c.Query("scene"); override != "" {
			sceneCode = override
		}
	}
	if sceneCode == "" {
		resolvedSceneCode, err := h.resolveSceneCode(c)
		if err != nil {
			h.respondSceneResolveError(c, err)
			return
		}
		sceneCode = resolvedSceneCode
	}

	result, err := h.contentService.GetBySceneAndDate(c.Request.Context(), sceneCode, date)
	if err != nil {
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
		return
	}

	model.WriteOK(c, result)
}

func (h *PublicContentHandler) GetSceneDomainMapping(c *gin.Context) {
	resolved, err := h.sceneResolver.ResolveHost(c.Request.Context(), middleware.RequestHost(c.Request))
	if err != nil {
		h.respondSceneResolveError(c, err)
		return
	}

	model.WriteOK(c, &model.PublicSceneDomainMappingResponse{
		Host:      resolved.Host,
		SceneCode: resolved.SceneCode,
	})
}

func (h *PublicContentHandler) GetScenePageConfig(c *gin.Context) {
	sceneCode, err := h.resolveSceneCode(c)
	if err != nil {
		h.respondSceneResolveError(c, err)
		return
	}

	result, err := h.scenePageConfigService.GetBySceneCode(c.Request.Context(), sceneCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidScenePageConfigParams):
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		case errors.Is(err, service.ErrScenePageConfigNotFound):
			model.WriteError(c, http.StatusNotFound, model.CodeScenePageConfigNotFound, "scene page config not found")
		default:
			model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
		}
		return
	}

	model.WriteOK(c, result)
}

func (h *PublicContentHandler) resolveSceneCode(c *gin.Context) (string, error) {
	resolved, err := h.sceneResolver.ResolveHost(c.Request.Context(), middleware.RequestHost(c.Request))
	if err != nil {
		return "", err
	}

	return resolved.SceneCode, nil
}

func (h *PublicContentHandler) respondSceneResolveError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSceneNotConfigured):
		model.WriteError(c, http.StatusNotFound, model.CodeSceneDomainNotFound, "scene not configured")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}
