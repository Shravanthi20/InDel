package middleware

import (
	"net"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiters = map[string]*rate.Limiter{}
var limitersMu sync.Mutex

func resetLimiters() {
	limitersMu.Lock()
	defer limitersMu.Unlock()
	limiters = map[string]*rate.Limiter{}
}

func clientIP(c *gin.Context) string {
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil && ip != "" {
		return ip
	}
	ip = c.ClientIP()
	if ip != "" {
		return ip
	}
	return c.Request.RemoteAddr
}

// RateLimitMiddleware enforces per-gateway rate limiting
func RateLimitMiddleware(requestsPerSecond float64) gin.HandlerFunc {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1
	}

	return func(c *gin.Context) {
		ip := clientIP(c)

		limitersMu.Lock()
		if limiters[ip] == nil {
			limiters[ip] = rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
		}
		limiter := limiters[ip]
		limitersMu.Unlock()

		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
