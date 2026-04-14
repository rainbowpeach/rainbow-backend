package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			model.ErrorResponse(model.CodeInternalServerError, "internal server error"),
		)
	})
}
