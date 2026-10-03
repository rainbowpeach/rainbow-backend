package handler

import (
	"errors"
	"log"
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

	sceneCode, err := h.resolveSceneCode(c)
	if err != nil {
		h.respondSceneResolveError(c, err)
		return
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
	requestHost := middleware.RequestHost(c.Request)
	normalizedHost := model.NormalizeHost(requestHost)
	forwardedHost := c.Request.Header.Get("X-Forwarded-Host")
	forwardedProto := c.Request.Header.Get("X-Forwarded-Proto")
	realIP := c.Request.Header.Get("X-Real-IP")
	origin := c.Request.Header.Get("Origin")
	referer := c.Request.Header.Get("Referer")
	userAgent := c.Request.Header.Get("User-Agent")
	sceneOverride := c.Query("scene")

	log.Printf(
		"public scene page config request ip=%s real_ip=%q host=%q x_forwarded_host=%q x_forwarded_proto=%q request_host=%q normalized_host=%q origin=%q referer=%q user_agent=%q scene_query=%q override_enabled=%t",
		c.ClientIP(),
		realIP,
		c.Request.Host,
		forwardedHost,
		forwardedProto,
		requestHost,
		normalizedHost,
		origin,
		referer,
		userAgent,
		sceneOverride,
		h.cfg.EnablePublicOverride,
	)

	sceneCode, err := h.resolveSceneCode(c)
	if err != nil {
		log.Printf(
			"public scene page config resolve failed ip=%s host=%q x_forwarded_host=%q request_host=%q normalized_host=%q origin=%q referer=%q scene_query=%q err=%v",
			c.ClientIP(),
			c.Request.Host,
			forwardedHost,
			requestHost,
			normalizedHost,
			origin,
			referer,
			sceneOverride,
			err,
		)
		h.respondSceneResolveError(c, err)
		return
	}

	result, err := h.scenePageConfigService.GetBySceneCode(c.Request.Context(), sceneCode)
	if err != nil {
		log.Printf(
			"public scene page config lookup failed ip=%s host=%q request_host=%q normalized_host=%q scene_code=%q err=%v",
			c.ClientIP(),
			c.Request.Host,
			requestHost,
			normalizedHost,
			sceneCode,
			err,
		)
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

	log.Printf(
		"public scene page config served ip=%s host=%q request_host=%q normalized_host=%q scene_code=%q",
		c.ClientIP(),
		c.Request.Host,
		requestHost,
		normalizedHost,
		sceneCode,
	)
	model.WriteOK(c, result)
}

func (h *PublicContentHandler) resolveSceneCode(c *gin.Context) (string, error) {
	if h.cfg.EnablePublicOverride {
		if override := c.Query("scene"); override != "" {
			return override, nil
		}
	}

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
