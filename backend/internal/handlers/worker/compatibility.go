package worker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetOrderDetail(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	orderID := c.Param("order_id")
	if strings.TrimSpace(orderID) == "" {
		c.JSON(400, gin.H{"error": "order_id_required"})
		return
	}

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			type row struct {
				ID         uint      `gorm:"column:id"`
				ZoneID     uint      `gorm:"column:zone_id"`
				OrderValue float64   `gorm:"column:order_value"`
				Status     string    `gorm:"column:status"`
				CreatedAt  time.Time `gorm:"column:created_at"`
			}
			var r row
			err := workerDB.Raw(`
				SELECT id, zone_id, order_value, status, created_at
				FROM orders
				WHERE worker_id = ? AND (CAST(id AS TEXT) = ? OR ? = CONCAT('ord-', LPAD(CAST(id AS TEXT), 3, '0')))
				ORDER BY id DESC
				LIMIT 1
			`, workerIDUint, orderID, orderID).Scan(&r).Error
			if err == nil && r.ID > 0 {
				c.JSON(200, gin.H{
					"order_id":      fmt.Sprintf("ord-%03d", r.ID),
					"worker_id":     workerIDUint,
					"zone_id":       r.ZoneID,
					"pickup_area":   "Pickup",
					"drop_area":     "Drop",
					"distance_km":   1.0,
					"earning_inr":   r.OrderValue,
					"status":        r.Status,
					"assigned_at":   r.CreatedAt.UTC().Format(time.RFC3339),
					"created_at":    r.CreatedAt.UTC().Format(time.RFC3339),
					"zone_level":    "A",
					"customer_name": "Customer",
				})
				return
			}
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, order := range store.data.Orders {
		id := bodyString(order, "order_id", "")
		if id == orderID {
			c.JSON(200, order)
			return
		}
	}

	c.JSON(404, gin.H{"error": "order_not_found"})
}

func SendCustomerCode(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	c.JSON(200, gin.H{"message": "customer_code_sent", "policy": gin.H{"status": "active"}})
}

func SendVerificationCode(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	c.JSON(200, gin.H{"message": "verification_code_sent", "policy": gin.H{"status": "active"}})
}

func VerifyFetchCode(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	code := bodyString(body, "code", "")
	if code == "" {
		c.JSON(400, gin.H{"error": "code_required"})
		return
	}
	c.JSON(200, gin.H{"message": "verification_successful", "policy": gin.H{"status": "active"}})
}

func GetZoneConfig(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	store.mu.RLock()
	profile := store.data.WorkerProfiles[workerID]
	store.mu.RUnlock()

	zoneName := bodyString(profile, "zone_name", "Tambaram")
	zoneID := bodyString(profile, "zone", "zone_1")
	if zoneID == "" {
		zoneID = "zone_1"
	}

	c.JSON(200, gin.H{
		"zone_id":               zoneID,
		"name":                  zoneName,
		"require_ip_validation": false,
	})
}

func GetSession(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	sessionID := c.Param("session_id")
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "session-demo"
	}
	c.JSON(200, gin.H{
		"session_id":           sessionID,
		"start_time":           time.Now().UTC().Add(-45 * time.Minute).Format(time.RFC3339),
		"end_time":             nil,
		"status":               "active",
		"deliveries_completed": 0,
		"earnings_in_session":  0,
	})
}

func GetSessionDeliveries(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	store.mu.RLock()
	orders := store.data.Orders
	store.mu.RUnlock()
	c.JSON(200, gin.H{"orders": orders})
}

func GetSessionFraudSignals(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	c.JSON(200, gin.H{"signals": []gin.H{}})
}

func EndSession(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	c.JSON(200, gin.H{"message": "session_ended", "policy": gin.H{"status": "active"}})
}

func DemoAssignOrders(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	count := bodyInt(body, "count", 3)
	if count < 1 {
		count = 1
	}
	c.JSON(200, gin.H{"message": "orders_assigned", "policy": gin.H{"status": "active"}, "count": strconv.Itoa(count)})
}

func DemoSimulateDeliveries(c *gin.Context) {
	_, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	count := bodyInt(body, "count", 3)
	if count < 1 {
		count = 1
	}
	c.JSON(200, gin.H{"message": "deliveries_simulated", "policy": gin.H{"status": "active"}, "count": strconv.Itoa(count)})
}
