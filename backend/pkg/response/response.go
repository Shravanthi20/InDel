package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Success returns a standard success response
func Success(c *gin.Context, statusCode int, data interface{}) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
	}

	c.JSON(statusCode, gin.H{
		"data": data,
		"meta": gin.H{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestID,
		},
	})
}

// Error returns a standard error response
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
