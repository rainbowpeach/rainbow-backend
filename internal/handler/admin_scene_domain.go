package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type AdminSceneDomainHandler struct {
	sceneDomainService *service.SceneDomainService
}

func NewAdminSceneDomainHandler(sceneDomainService *service.SceneDomainService) *AdminSceneDomainHandler {
	return &AdminSceneDomainHandler{sceneDomainService: sceneDomainService}
}

func (h *AdminSceneDomainHandler) List(c *gin.Context) {
	var req model.SceneDomainListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("admin scene domain list invalid query %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.sceneDomainService.List(c.Request.Context(), req)
	if err != nil {
		h.respondSceneDomainError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminSceneDomainHandler) GetByHost(c *gin.Context) {
	host, ok := sceneDomainHostParam(c)
	if !ok {
		return
	}

	result, err := h.sceneDomainService.GetByHost(c.Request.Context(), host)
	if err != nil {
		h.respondSceneDomainError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminSceneDomainHandler) Create(c *gin.Context) {
	var req model.SceneDomainUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin scene domain create invalid request %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.sceneDomainService.Create(c.Request.Context(), &req)
	if err != nil {
		h.respondSceneDomainError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminSceneDomainHandler) Update(c *gin.Context) {
	host, ok := sceneDomainHostParam(c)
	if !ok {
		return
	}

	var req model.SceneDomainUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin scene domain update invalid request %s ip=%s host=%q err=%v", adminActor(c), c.ClientIP(), host, err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	if strings.TrimSpace(req.Host) == "" {
		req.Host = host
	}

	result, err := h.sceneDomainService.Update(c.Request.Context(), host, &model.SceneDomainUpsertRequest{
		Host:      req.Host,
		SceneCode: req.SceneCode,
	})
	if err != nil {
		h.respondSceneDomainError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminSceneDomainHandler) Delete(c *gin.Context) {
	host, ok := sceneDomainHostParam(c)
	if !ok {
		return
	}

	result, err := h.sceneDomainService.Delete(c.Request.Context(), host)
	if err != nil {
		h.respondSceneDomainError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminSceneDomainHandler) respondSceneDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidContentParams):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrDuplicateHost):
		model.WriteError(c, http.StatusBadRequest, model.CodeDuplicateHost, "duplicate host")
	case errors.Is(err, service.ErrSceneDomainNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeSceneDomainNotFound, "scene domain not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}

func sceneDomainHostParam(c *gin.Context) (string, bool) {
	rawValue := c.Param("host")
	value, err := url.PathUnescape(rawValue)
	if err != nil {
		log.Printf("invalid scene domain host param value=%q ip=%s path=%s err=%v", rawValue, c.ClientIP(), c.Request.URL.Path, err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return "", false
	}

	if value == "" {
		log.Printf("empty scene domain host param ip=%s path=%s", c.ClientIP(), c.Request.URL.Path)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return "", false
	}

	return value, true
}
