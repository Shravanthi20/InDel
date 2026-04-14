package worker

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func notificationTitle(kind string) string {
	switch kind {
	case "disruption_alert":
		return "Disruption detected"
	case "claim_generated":
		return "Claim Generated"
	case "payout_credited":
		return "Payout credited"
	case "order_delivered":
		return "Order delivered"
	case "premium_due":
		return "Premium updated"
	default:
		return "Notification"
	}
}

// GetNotifications returns notifications
func GetNotifications(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type row struct {
				ID        uint   `gorm:"column:id"`
				Type      string `gorm:"column:type"`
				Message   string `gorm:"column:message"`
				ReadAt    string `gorm:"column:read_at"`
				CreatedAt string `gorm:"column:created_at"`
			}
			rows := make([]row, 0)
			_ = workerDB.Raw(
				"SELECT id, type, message, COALESCE(read_at::text, '') as read_at, created_at::text FROM notifications WHERE worker_id = ? ORDER BY created_at DESC LIMIT 50",
				workerIDUint,
			).Scan(&rows).Error

			notifications := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				notifications = append(notifications, gin.H{
					"id":         fmt.Sprintf("ntf_%d", row.ID),
					"type":       row.Type,
					"title":      notificationTitle(row.Type),
					"body":       row.Message,
					"created_at": row.CreatedAt,
					"read":       row.ReadAt != "",
				})
			}
			c.JSON(200, gin.H{"notifications": notifications})
			return
		}
	}

	store.mu.RLock()
	notifications := store.data.Notifications
	store.mu.RUnlock()

	c.JSON(200, gin.H{"notifications": notifications})
}

// SetNotificationPreferences sets user preferences
func SetNotificationPreferences(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	body := parseBody(c)
	c.JSON(200, gin.H{"preferences": body})
}

// RegisterFCMToken registers FCM token
func RegisterFCMToken(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	token := bodyString(body, "fcm_token", "")
	if token == "" {
		c.JSON(400, gin.H{"registered": false, "error": "fcm_token_required"})
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`INSERT INTO fcm_tokens (worker_id, token, device_name)
				 VALUES (?, ?, ?)
				 ON CONFLICT (token) DO UPDATE SET worker_id = EXCLUDED.worker_id, device_name = EXCLUDED.device_name`,
				workerIDUint, token, bodyString(body, "device_name", "android"),
			).Error
		}
	}
	c.JSON(200, gin.H{"registered": token != ""})
}
