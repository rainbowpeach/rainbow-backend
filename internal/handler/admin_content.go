package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type AdminContentHandler struct {
	contentService *service.ContentService
}

func NewAdminContentHandler(contentService *service.ContentService) *AdminContentHandler {
	return &AdminContentHandler{
		contentService: contentService,
	}
}

func (h *AdminContentHandler) Create(c *gin.Context) {
	var req model.ContentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin content create invalid request %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.contentService.Create(c.Request.Context(), &req)
	if err != nil {
		log.Printf("admin content create failed %s ip=%s scene=%s date=%s err=%v", adminActor(c), c.ClientIP(), req.SceneCode, req.Date, err)
		h.respondContentError(c, err)
		return
	}

	log.Printf("admin content created %s ip=%s content_id=%d scene=%s date=%s", adminActor(c), c.ClientIP(), result.ID, req.SceneCode, req.Date)
	model.WriteOK(c, result)
}

func (h *AdminContentHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req model.ContentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("admin content update invalid request %s ip=%s content_id=%d err=%v", adminActor(c), c.ClientIP(), id, err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.contentService.Update(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("admin content update failed %s ip=%s content_id=%d scene=%s date=%s err=%v", adminActor(c), c.ClientIP(), id, req.SceneCode, req.Date, err)
		h.respondContentError(c, err)
		return
	}

	log.Printf("admin content updated %s ip=%s content_id=%d scene=%s date=%s", adminActor(c), c.ClientIP(), result.ID, req.SceneCode, req.Date)
	model.WriteOK(c, result)
}

func (h *AdminContentHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	result, err := h.contentService.Delete(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin content delete failed %s ip=%s content_id=%d err=%v", adminActor(c), c.ClientIP(), id, err)
		h.respondContentError(c, err)
		return
	}

	log.Printf("admin content deleted %s ip=%s content_id=%d", adminActor(c), c.ClientIP(), result.ID)
	model.WriteOK(c, result)
}

func (h *AdminContentHandler) List(c *gin.Context) {
	var req model.ContentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("admin content list invalid query %s ip=%s err=%v", adminActor(c), c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.contentService.List(c.Request.Context(), model.ContentFilter{
		SceneCode: req.Scene,
		Date:      req.Date,
	}, req.Page, req.PageSize)
	if err != nil {
		log.Printf("admin content list failed %s ip=%s page=%d page_size=%d scene=%s date=%s err=%v", adminActor(c), c.ClientIP(), req.Page, req.PageSize, req.Scene, req.Date, err)
		h.respondContentError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *AdminContentHandler) respondContentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidContentParams):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrInvalidDateFormat):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidDateFormat, "invalid date format")
	case errors.Is(err, service.ErrDuplicateDate):
		model.WriteError(c, http.StatusBadRequest, model.CodeDuplicateDate, "duplicate date")
	case errors.Is(err, service.ErrContentNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeContentNotFound, "content not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	value := c.Param(key)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		log.Printf("invalid uint param key=%s value=%q ip=%s path=%s", key, value, c.ClientIP(), c.Request.URL.Path)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return 0, false
	}

	return uint(id), true
}
