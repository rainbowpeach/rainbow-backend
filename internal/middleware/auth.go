package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
	"rainbow-backend/internal/service"
)

const (
	ContextAdminIDKey  = "adminID"
	ContextUsernameKey = "adminUsername"
	ContextUserIDKey   = "userID"
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

func UserTokenAuth(users repo.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			log.Printf("user authorization header missing or invalid ip=%s path=%s", c.ClientIP(), c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			log.Printf("user token empty ip=%s path=%s", c.ClientIP(), c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		user, err := users.GetByToken(c.Request.Context(), token)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("user token lookup failed ip=%s path=%s err=%v", c.ClientIP(), c.Request.URL.Path, err)
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}
		if user.TokenExpireAt == nil || !user.TokenExpireAt.After(time.Now()) {
			log.Printf("user token expired ip=%s path=%s user_id=%d", c.ClientIP(), c.Request.URL.Path, user.ID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		c.Set(ContextUserIDKey, user.ID)
		c.Next()
	}
}

func UserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint)
	return userID, ok
}
