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
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidParams, "invalid params"))
		return
	}

	result, err := h.contentService.GetByDate(c.Request.Context(), date)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDateFormat):
			c.JSON(http.StatusBadRequest, model.ErrorResponse(model.CodeInvalidDateFormat, "invalid date format"))
		case errors.Is(err, service.ErrContentNotFound):
			c.JSON(http.StatusNotFound, model.ErrorResponse(model.CodeContentNotFound, "content not found"))
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.CodeInternalServerError, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}
