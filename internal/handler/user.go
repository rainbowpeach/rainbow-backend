package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/middleware"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

type UserHandler struct {
	profileService *service.UserProfileService
	uploadService  *service.UploadService
}

func NewUserHandler(profileService *service.UserProfileService, uploadService *service.UploadService) *UserHandler {
	return &UserHandler{
		profileService: profileService,
		uploadService:  uploadService,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	result, err := h.profileService.Get(c.Request.Context(), userID)
	if err != nil {
		log.Printf("user profile get failed user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		h.respondProfileError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *UserHandler) SaveProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req model.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("user profile save invalid request user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
		return
	}

	result, err := h.profileService.Save(c.Request.Context(), userID, &req)
	if err != nil {
		log.Printf("user profile save failed user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		h.respondProfileError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	result, err := h.profileService.Clear(c.Request.Context(), userID)
	if err != nil {
		log.Printf("user profile clear failed user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		h.respondProfileError(c, err)
		return
	}

	model.WriteOK(c, result)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			log.Printf("user avatar upload missing file user_id=%d ip=%s", userID, c.ClientIP())
			model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "file is required")
			return
		}
		log.Printf("user avatar upload invalid multipart form user_id=%d ip=%s err=%v", userID, c.ClientIP(), err)
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid multipart form")
		return
	}

	result, err := h.uploadService.UploadAvatar(c.Request.Context(), &service.UploadRequest{
		FileHeader: fileHeader,
		BaseURL:    middleware.RequestBaseURL(c.Request),
	})
	if err != nil {
		log.Printf(
			"user avatar upload failed user_id=%d ip=%s filename=%q size=%d err=%v",
			userID,
			c.ClientIP(),
			fileHeader.Filename,
			fileHeader.Size,
			err,
		)
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
		"user avatar upload succeeded user_id=%d ip=%s filename=%q stored=%q size=%d content_type=%s",
		userID,
		c.ClientIP(),
		fileHeader.Filename,
		result.Filename,
		result.Size,
		result.ContentType,
	)
	model.WriteOK(c, result)
}

func (h *UserHandler) respondProfileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUserProfile):
		model.WriteError(c, http.StatusBadRequest, model.CodeInvalidParams, "invalid params")
	case errors.Is(err, service.ErrUserNotFound):
		model.WriteError(c, http.StatusNotFound, model.CodeUserNotFound, "user not found")
	default:
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
	}
}

func currentUserID(c *gin.Context) (uint, bool) {
	userID, ok := middleware.UserID(c)
	if ok {
		return userID, true
	}

	model.WriteError(c, http.StatusUnauthorized, model.CodeUnauthorized, "unauthorized")
	return 0, false
}
