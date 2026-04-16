package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/middleware"
)

func adminActor(c *gin.Context) string {
	adminID, hasAdminID := c.Get(middleware.ContextAdminIDKey)
	username, hasUsername := c.Get(middleware.ContextUsernameKey)
	if !hasAdminID && !hasUsername {
		return "admin_id=unknown username=unknown"
	}

	return fmt.Sprintf("admin_id=%v username=%v", adminID, username)
}
