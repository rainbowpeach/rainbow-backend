package middleware

import (
	"log"
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
			log.Printf("jwt authorization header missing or invalid ip=%s path=%s", c.ClientIP(), c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			log.Printf("jwt token empty ip=%s path=%s", c.ClientIP(), c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		claims, err := tokens.Parse(token)
		if err != nil {
			log.Printf("jwt token parse failed ip=%s path=%s err=%v", c.ClientIP(), c.Request.URL.Path, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		c.Set(ContextAdminIDKey, claims.AdminID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}
