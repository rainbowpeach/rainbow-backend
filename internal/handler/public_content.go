package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type PublicContentHandler struct {
	contentService *service.ContentService
}

func NewPublicContentHandler(contentService *service.ContentService) *PublicContentHandler {
	return &PublicContentHandler{
		contentService: contentService,
	}
}

func (h *PublicContentHandler) GetByDate(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.contentService.GetByDate(c.Request.Context(), date)
	if err != nil {
		switch {
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
