package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v", recovered)
		model.WriteError(c, http.StatusInternalServerError, model.CodeInternalServerError, "internal server error")
		c.Abort()
	})
}
