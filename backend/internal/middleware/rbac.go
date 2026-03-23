package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RBACMiddleware for role-based access control
func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if len(allowed) == 0 {
			c.Next()
			return
		}

		role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
		if role == "" {
			if rawClaims, ok := c.Get("claims"); ok {
				if claims, ok := rawClaims.(jwt.MapClaims); ok {
					if claimRole, ok := claims["role"].(string); ok {
						role = strings.ToLower(strings.TrimSpace(claimRole))
					}
				}
			}
		}

		if role == "" {
			c.JSON(403, gin.H{"error": "missing role"})
			c.Abort()
			return
		}

		if _, ok := allowed[role]; !ok {
			c.JSON(403, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
