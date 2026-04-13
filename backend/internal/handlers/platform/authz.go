package platform

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequirePlatformOperatorRole protects platform control endpoints (demo actions).
func RequirePlatformOperatorRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedRoles := allowedRolesFromEnv("INDEL_PLATFORM_OPERATOR_ALLOWED_ROLES", []string{"admin", "platform_admin", "ops_manager"})
		role, ok := roleFromBearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_or_invalid_bearer_token"})
			return
		}
		if !containsRole(allowedRoles, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "forbidden_platform_operation",
				"role":          role,
				"allowed_roles": allowedRoles,
			})
			return
		}
		c.Next()
	}
}

// RequirePlatformWebhookAuth protects platform webhook endpoints via role or shared key.
func RequirePlatformWebhookAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasValidPlatformWebhookKey(c.GetHeader("X-Platform-Webhook-Key")) {
			c.Next()
			return
		}

		allowedRoles := allowedRolesFromEnv("INDEL_PLATFORM_WEBHOOK_ALLOWED_ROLES", []string{"admin", "platform_admin", "ops_manager"})
		role, ok := roleFromBearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_or_invalid_webhook_auth"})
			return
		}
		if !containsRole(allowedRoles, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "forbidden_platform_webhook",
				"role":          role,
				"allowed_roles": allowedRoles,
			})
			return
		}
		c.Next()
	}
}

func hasValidPlatformWebhookKey(provided string) bool {
	expected := strings.TrimSpace(os.Getenv("INDEL_PLATFORM_WEBHOOK_KEY"))
	provided = strings.TrimSpace(provided)
	if expected == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func roleFromBearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" || platformDB == nil {
		return "", false
	}

	type roleRow struct {
		Role string `gorm:"column:role"`
	}
	var row roleRow
	err := platformDB.Raw(
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

func allowedRolesFromEnv(envKey string, defaults []string) []string {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return defaults
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
		return defaults
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
