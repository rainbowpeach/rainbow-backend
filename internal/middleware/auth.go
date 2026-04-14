package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/service"
)

const (
	ContextAdminIDKey  = "adminID"
	ContextUsernameKey = "adminUsername"
)

func JWTAuth(tokens service.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		claims, err := tokens.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		c.Set(ContextAdminIDKey, claims.AdminID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}
