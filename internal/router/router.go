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
	registerStaticRoutes(engine, cfg.Upload.RootDir)

	adminRepo := repo.NewAdminRepository(db)
	contentRepo := repo.NewContentRepository(db)
	sceneDomainRepo := repo.NewSceneDomainRepository(db)
	scenePageConfigRepo := repo.NewScenePageConfigRepository(db)
	tokenManager := service.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	authService := service.NewAuthService(adminRepo, tokenManager)
	contentService := service.NewContentService(contentRepo)
	sceneDomainService := service.NewSceneDomainService(sceneDomainRepo)
	scenePageConfigService := service.NewScenePageConfigService(scenePageConfigRepo)
	sceneResolver := service.NewSceneResolver(sceneDomainRepo)
	uploadService := service.NewUploadService(cfg.Upload)
	adminAuthHandler := handler.NewAdminAuthHandler(authService)
	adminContentHandler := handler.NewAdminContentHandler(contentService)
	adminUploadHandler := handler.NewAdminUploadHandler(uploadService)
	adminSceneDomainHandler := handler.NewAdminSceneDomainHandler(sceneDomainService)
	adminScenePageConfigHandler := handler.NewAdminScenePageConfigHandler(scenePageConfigService)
	publicContentHandler := handler.NewPublicContentHandler(contentService, scenePageConfigService, sceneResolver, cfg.Scene)
	engine.GET("/health", healthHandler(cfg, db))

	public := engine.Group("/api/public")
	public.GET("/content", publicContentHandler.GetByDate)
	public.GET("/scene-domain-mapping", publicContentHandler.GetSceneDomainMapping)
	public.GET("/scene-page-config", publicContentHandler.GetScenePageConfig)

	admin := engine.Group("/api/admin")
	admin.POST("/login", adminAuthHandler.Login)

	adminProtected := admin.Group("")
	adminProtected.Use(middleware.JWTAuth(tokenManager))
	adminProtected.POST("/content", adminContentHandler.Create)
	adminProtected.PUT("/content/:id", adminContentHandler.Update)
	adminProtected.DELETE("/content/:id", adminContentHandler.Delete)
	adminProtected.GET("/content", adminContentHandler.List)
	adminProtected.GET("/scene-domains", adminSceneDomainHandler.List)
	adminProtected.GET("/scene-domains/:host", adminSceneDomainHandler.GetByHost)
	adminProtected.POST("/scene-domains", adminSceneDomainHandler.Create)
	adminProtected.PUT("/scene-domains/:host", adminSceneDomainHandler.Update)
	adminProtected.DELETE("/scene-domains/:host", adminSceneDomainHandler.Delete)
	adminProtected.GET("/scene-page-configs", adminScenePageConfigHandler.List)
	adminProtected.GET("/scene-page-configs/:scene_code", adminScenePageConfigHandler.GetBySceneCode)
	adminProtected.POST("/scene-page-configs", adminScenePageConfigHandler.Create)
	adminProtected.PUT("/scene-page-configs/:scene_code", adminScenePageConfigHandler.Update)
	adminProtected.DELETE("/scene-page-configs/:scene_code", adminScenePageConfigHandler.Delete)
	adminProtected.POST("/upload/image", adminUploadHandler.UploadImage)
	adminProtected.POST("/upload/audio", adminUploadHandler.UploadAudio)

	return engine
}

func registerStaticRoutes(engine *gin.Engine, uploadRoot string) {
	engine.StaticFS("/static", gin.Dir(uploadRoot, false))
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
