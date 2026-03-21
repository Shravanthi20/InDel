package middleware

import (
	"net"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiters = map[string]*rate.Limiter{}

// RateLimitMiddleware enforces per-gateway rate limiting
func RateLimitMiddleware(requestsPerSecond float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

		if limiters[ip] == nil {
			limiters[ip] = rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
		}

		if !limiters[ip].Allow() {
			c.JSON(429, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
