package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return fmt.Sprintf(
			"%s | %3d | %13v | %15s | %-7s %s\n",
			params.TimeStamp.Format(time.RFC3339),
			params.StatusCode,
			params.Latency,
			params.ClientIP,
			params.Method,
			params.Path,
		)
	})
}
