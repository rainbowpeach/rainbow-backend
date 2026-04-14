package handler

import (
	"errors"
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
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return
	}

	result, err := h.contentService.Create(c.Request.Context(), &req)
	if err != nil {
		h.respondContentError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AdminContentHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req model.ContentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return
	}

	result, err := h.contentService.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.respondContentError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AdminContentHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	result, err := h.contentService.Delete(c.Request.Context(), id)
	if err != nil {
		h.respondContentError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AdminContentHandler) List(c *gin.Context) {
	var req model.ContentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return
	}

	result, err := h.contentService.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.CodeInternalServerError, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AdminContentHandler) respondContentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidDateFormat, "invalid date format"))
	case errors.Is(err, service.ErrDuplicateDate):
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeDuplicateDate, "duplicate date"))
	case errors.Is(err, service.ErrContentNotFound):
		c.JSON(http.StatusNotFound, model.ErrorResponse(model.CodeContentNotFound, "content not found"))
	default:
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.CodeInternalServerError, "internal server error"))
	}
}

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	value := c.Param(key)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return 0, false
	}

	return uint(id), true
}
