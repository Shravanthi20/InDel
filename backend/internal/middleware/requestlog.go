package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogMiddleware logs all API requests
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime).Milliseconds()
		_ = duration
	}
}
