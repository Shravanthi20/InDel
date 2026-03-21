package response

import "github.com/gin-gonic/gin"

// Success returns a standard success response
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"data": data,
		"meta": gin.H{
			"timestamp":  "",
			"request_id": "",
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
