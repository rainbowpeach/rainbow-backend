package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/middleware"
	"rainbow-backend/internal/model"
)

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS(cfg.AllowOrigins))

	engine.GET("/health", healthHandler(cfg, db))

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
