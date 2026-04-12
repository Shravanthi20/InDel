package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS is a simple middleware to handle cross-origin requests.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := resolveAllowedOrigin(origin)
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		} else if strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		// Needed by newer browsers for cross-origin requests from local pages to private network hosts.
		c.Writer.Header().Set("Access-Control-Allow-Private-Network", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func resolveAllowedOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	configured := strings.TrimSpace(os.Getenv("INDEL_ALLOWED_ORIGINS"))
	if configured == "" {
		return ""
	}

	for _, item := range strings.Split(configured, ",") {
		candidate := strings.TrimSpace(item)
		if candidate == "" {
			continue
		}
		if candidate == "*" || strings.EqualFold(candidate, origin) {
			return origin
		}
	}

	return ""
}
