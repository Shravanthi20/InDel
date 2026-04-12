package core

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireInternalRole guards core internal endpoints with role-based authorization.
func RequireInternalRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedRoles := allowedCoreRolesFromEnv()

		role, ok := roleFromBearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_or_invalid_bearer_token"})
			return
		}

		if !containsRole(allowedRoles, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "forbidden_internal_operation",
				"role":          role,
				"allowed_roles": allowedRoles,
			})
			return
		}

		c.Next()
	}
}

func roleFromBearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" || coreDB == nil {
		return "", false
	}

	type roleRow struct {
		Role string `gorm:"column:role"`
	}
	var row roleRow
	err := coreDB.Raw(
		`SELECT COALESCE(u.role, '') AS role
		 FROM auth_tokens t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.token = ? AND t.expires_at > CURRENT_TIMESTAMP
		 LIMIT 1`,
		token,
	).Scan(&row).Error
	if err != nil {
		return "", false
	}

	role := strings.ToLower(strings.TrimSpace(row.Role))
	if role == "" {
		return "", false
	}
	return role, true
}

func allowedCoreRolesFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("INDEL_CORE_INTERNAL_ALLOWED_ROLES"))
	if raw == "" {
		return []string{"admin", "platform_admin", "ops_manager"}
	}
	parts := strings.Split(raw, ",")
	roles := make([]string, 0, len(parts))
	for _, p := range parts {
		role := strings.ToLower(strings.TrimSpace(p))
		if role != "" {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		return []string{"admin", "platform_admin", "ops_manager"}
	}
	return roles
}

func containsRole(roles []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, role := range roles {
		if strings.ToLower(strings.TrimSpace(role)) == target {
			return true
		}
	}
	return false
}
