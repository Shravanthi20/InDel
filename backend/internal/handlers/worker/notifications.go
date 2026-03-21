package worker

import "github.com/gin-gonic/gin"

// GetNotifications returns notifications
func GetNotifications(c *gin.Context) {
	// GET /api/notifications
	c.JSON(200, gin.H{"notifications": []interface{}{}})
}

// SetNotificationPreferences sets user preferences
func SetNotificationPreferences(c *gin.Context) {
	// PUT /api/notifications/preferences
	c.JSON(200, gin.H{"preferences": map[string]bool{}})
}

// RegisterFCMToken registers FCM token
func RegisterFCMToken(c *gin.Context) {
	// POST /api/notifications/fcm-token
	c.JSON(200, gin.H{"registered": true})
}
