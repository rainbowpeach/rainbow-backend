package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type AdminScenePageConfigHandler struct {
	scenePageConfigService *service.ScenePageConfigService
}

func NewAdminScenePageConfigHandler(scenePageConfigService *service.ScenePageConfigService) *AdminScenePageConfigHandler {
	return &AdminScenePageConfigHandler{scenePageConfigService: scenePageConfigService}
}

func (h *AdminScenePageConfigHandler) List(c *gin.Context) {
	var req model.ScenePageConfigListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("admin scene page config list invalid query %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.scenePageConfigService.List(c.Request.Context(), req)
	if err != nil {
		h.respondScenePageConfigError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminScenePageConfigHandler) GetBySceneCode(c *gin.Context) {
	sceneCode := c.Param("scene_code")
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

func (h *AdminScenePageConfigHandler) Create(c *gin.Context) {
	var req model.ScenePageConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin scene page config create invalid request %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.scenePageConfigService.Create(c.Request.Context(), &req)
	if err != nil {
		log.Printf("admin scene page config create failed %s ip=%s scene=%s err=%v", adminActor(c), c.ClientIP(), req.SceneCode, err)
		h.respondScenePageConfigError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminScenePageConfigHandler) Update(c *gin.Context) {
	sceneCode := c.Param("scene_code")
	if sceneCode == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	var req model.ScenePageConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin scene page config update invalid request %s ip=%s scene=%s err=%v", adminActor(c), c.ClientIP(), sceneCode, err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}
	if req.SceneCode == "" {
		req.SceneCode = sceneCode
	}

	result, err := h.scenePageConfigService.Update(c.Request.Context(), sceneCode, &req)
	if err != nil {
		log.Printf("admin scene page config update failed %s ip=%s scene=%s err=%v", adminActor(c), c.ClientIP(), sceneCode, err)
		h.respondScenePageConfigError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminScenePageConfigHandler) Delete(c *gin.Context) {
	sceneCode := c.Param("scene_code")
	if sceneCode == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.scenePageConfigService.Delete(c.Request.Context(), sceneCode)
	if err != nil {
		log.Printf("admin scene page config delete failed %s ip=%s scene=%s err=%v", adminActor(c), c.ClientIP(), sceneCode, err)
		h.respondScenePageConfigError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminScenePageConfigHandler) respondScenePageConfigError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidScenePageConfigParams):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrDuplicateScenePageConfig):
		model.WriteError(c, http.StatusBadRequest, model.CodeDuplicateScenePageConfig, "duplicate scene_code")
	case errors.Is(err, service.ErrScenePageConfigNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeScenePageConfigNotFound, "scene page config not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}
