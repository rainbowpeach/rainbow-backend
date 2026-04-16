package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		path := params.Path
		if params.Request != nil && params.Request.URL != nil && params.Request.URL.RawQuery != "" {
			path = path + "?" + params.Request.URL.RawQuery
		}

		errorMessage := ""
		if params.ErrorMessage != "" {
			errorMessage = " | " + strings.TrimSpace(params.ErrorMessage)
		}

		return fmt.Sprintf(
			"%s | %3d | %13v | %15s | %-7s %s%s\n",
			params.TimeStamp.Format(time.RFC3339),
			params.StatusCode,
			params.Latency,
			params.ClientIP,
			params.Method,
			path,
			errorMessage,
		)
	})
}
