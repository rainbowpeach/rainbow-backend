package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/handler"
	"rainbow-backend/internal/middleware"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
	"rainbow-backend/internal/service"
)

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS(cfg.AllowOrigins))

	adminRepo := repo.NewAdminRepository(db)
	contentRepo := repo.NewContentRepository(db)
	tokenManager := service.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	authService := service.NewAuthService(adminRepo, tokenManager)
	contentService := service.NewContentService(contentRepo)
	adminAuthHandler := handler.NewAdminAuthHandler(authService)
	adminContentHandler := handler.NewAdminContentHandler(contentService)
	publicContentHandler := handler.NewPublicContentHandler(contentService)

	engine.GET("/health", healthHandler(cfg, db))

	public := engine.Group("/api/public")
	public.GET("/content", publicContentHandler.GetByDate)

	admin := engine.Group("/api/admin")
	admin.POST("/login", adminAuthHandler.Login)

	adminProtected := admin.Group("")
	adminProtected.Use(middleware.JWTAuth(tokenManager))
	adminProtected.POST("/content", adminContentHandler.Create)
	adminProtected.PUT("/content/:id", adminContentHandler.Update)
	adminProtected.DELETE("/content/:id", adminContentHandler.Delete)
	adminProtected.GET("/content", adminContentHandler.List)

	return engine
}

func healthHandler(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				model.ErrorResponse(model.CodeInternalServerError, "database unavailable"),
			)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(
				http.StatusServiceUnavailable,
				model.ErrorResponse(model.CodeInternalServerError, "database unavailable"),
			)
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse(gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
		}))
	}
}
