package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type AdminUploadHandler struct {
	uploadService *service.UploadService
}

func NewAdminUploadHandler(uploadService *service.UploadService) *AdminUploadHandler {
	return &AdminUploadHandler{uploadService: uploadService}
}

func (h *AdminUploadHandler) UploadImage(c *gin.Context) {
	h.upload(c, h.uploadService.UploadImage)
}

func (h *AdminUploadHandler) UploadAudio(c *gin.Context) {
	h.upload(c, h.uploadService.UploadAudio)
}

func (h *AdminUploadHandler) upload(c *gin.Context, uploadFn func(context.Context, *service.UploadRequest) (*model.UploadResponse, error)) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			log.Printf("admin upload missing file %s ip=%s path=%s", adminActor(c), c.ClientIP(), c.Request.URL.Path)
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "file is required")
			return
		}
		log.Printf("admin upload invalid multipart form %s ip=%s path=%s err=%v", adminActor(c), c.ClientIP(), c.Request.URL.Path, err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid multipart form")
		return
	}

	result, err := uploadFn(c.Request.Context(), &service.UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    requestBaseURL(c.Request),
	})
	if err != nil {
		log.Printf("admin upload failed %s ip=%s path=%s filename=%q size=%d err=%v", adminActor(c), c.ClientIP(), c.Request.URL.Path, fileHeader.Filename, fileHeader.Size, err)
		switch {
		case errors.Is(err, service.ErrFileRequired):
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "file is required")
		case errors.Is(err, service.ErrEmptyFile):
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "empty file")
		case errors.Is(err, service.ErrUnsupportedFileType):
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "unsupported file type")
		case errors.Is(err, service.ErrFileTooLarge):
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "file too large")
		default:
			model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
		}
		return
	}

	log.Printf(
		"admin upload succeeded %s ip=%s path=%s filename=%q stored=%q size=%d content_type=%s",
		adminActor(c),
		c.ClientIP(),
		c.Request.URL.Path,
		fileHeader.Filename,
		result.Filename,
		result.Size,
		result.ContentType,
	)
	model.WriteOK(c, result)
}

func requestBaseURL(r *http.Request) string {
	scheme := forwardedHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := forwardedHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}

	return scheme + "://" + host
}

func forwardedHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}
