package worker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseOrderID(orderID string) (uint, error) {
	trimmed := strings.TrimSpace(orderID)
	trimmed = strings.TrimPrefix(trimmed, "ord-")
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

// GetAssignedOrders returns assigned orders only.
func GetAssignedOrders(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type orderRow struct {
				ID         uint    `gorm:"column:id"`
				OrderValue float64 `gorm:"column:order_value"`
				CreatedAt  string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			_ = workerDB.Raw("SELECT id, order_value, created_at::text FROM orders WHERE worker_id = ? ORDER BY created_at DESC LIMIT 20", workerIDUint).Scan(&rows).Error
			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				orders = append(orders, gin.H{
					"order_id":    fmt.Sprintf("ord-%03d", row.ID),
					"pickup_area": "Tambaram",
					"drop_area":   "Camp Road",
					"distance_km": 3.1,
					"earning_inr": int(row.OrderValue),
					"status":      "assigned",
					"assigned_at": row.CreatedAt,
				})
			}
			c.JSON(200, gin.H{"orders": orders})
			return
		}
	}

	store.mu.RLock()
	assigned := make([]map[string]any, 0)
	for _, order := range store.data.Orders {
		if order["status"] == "assigned" {
			assigned = append(assigned, order)
		}
	}
	store.mu.RUnlock()

	c.JSON(200, gin.H{"orders": assigned})
}

// GetOrders returns all worker orders.
func GetOrders(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type orderRow struct {
				ID         uint    `gorm:"column:id"`
				OrderValue float64 `gorm:"column:order_value"`
				CreatedAt  string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			_ = workerDB.Raw("SELECT id, order_value, created_at::text FROM orders WHERE worker_id = ? ORDER BY created_at DESC LIMIT 50", workerIDUint).Scan(&rows).Error
			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				orders = append(orders, gin.H{
					"order_id":    fmt.Sprintf("ord-%03d", row.ID),
					"pickup_area": "Tambaram",
					"drop_area":   "Camp Road",
					"distance_km": 3.1,
					"earning_inr": int(row.OrderValue),
					"status":      "assigned",
					"assigned_at": row.CreatedAt,
				})
			}
			c.JSON(200, gin.H{"orders": orders})
			return
		}
	}

	store.mu.RLock()
	orders := append([]map[string]any{}, store.data.Orders...)
	store.mu.RUnlock()

	c.JSON(200, gin.H{"orders": orders})
}

func updateOrderStatus(c *gin.Context, newStatus string, message string) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	orderID := c.Param("order_id")

	if hasDB() {
		workerIDUint, parseWorkerErr := parseWorkerID(workerID)
		orderNumID, parseOrderErr := parseOrderID(orderID)
		if parseWorkerErr == nil && parseOrderErr == nil {
			if newStatus == "accepted" {
				_ = workerDB.Exec("UPDATE orders SET status='accepted', accepted_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Error
			}
			if newStatus == "picked_up" {
				_ = workerDB.Exec("UPDATE orders SET status='picked_up', picked_up_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Error
			}
			if newStatus == "delivered" {
				_ = workerDB.Exec("UPDATE orders SET status='delivered', delivered_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Error
				_ = workerDB.Exec("INSERT INTO notifications (worker_id, type, message) VALUES (?, 'order_delivered', ?)", workerIDUint, fmt.Sprintf("%s delivered. Earnings updated.", orderID)).Error
			}

			type row struct {
				ID         uint    `gorm:"column:id"`
				OrderValue float64 `gorm:"column:order_value"`
				Status     string  `gorm:"column:status"`
				CreatedAt  string  `gorm:"column:created_at"`
				UpdatedAt  string  `gorm:"column:updated_at"`
			}
			var r row
			err := workerDB.Raw("SELECT id, order_value, status, created_at::text, updated_at::text FROM orders WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Scan(&r).Error
			if err == nil && r.ID != 0 {
				c.JSON(200, gin.H{"message": message, "order": gin.H{
					"order_id":    fmt.Sprintf("ord-%03d", r.ID),
					"pickup_area": "Tambaram",
					"drop_area":   "Camp Road",
					"distance_km": 3.1,
					"earning_inr": int(r.OrderValue),
					"status":      r.Status,
					"assigned_at": r.CreatedAt,
					"updated_at":  r.UpdatedAt,
				}})
				return
			}
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for _, order := range store.data.Orders {
		if order["order_id"] != orderID {
			continue
		}

		previousStatus, _ := order["status"].(string)
		order["status"] = newStatus
		order["updated_at"] = nowISO()

		if newStatus == "delivered" && previousStatus != "delivered" {
			earning, _ := order["earning_inr"].(int)
			if current, ok := store.data.Earnings["this_week_actual"].(int); ok {
				store.data.Earnings["this_week_actual"] = current + earning
			}
			if profile, exists := store.data.WorkerProfiles[workerID]; exists {
				if completed, ok := profile["orders_completed"].(int); ok {
					profile["orders_completed"] = completed + 1
				}
			}

			store.data.Notifications = append([]map[string]any{{
				"id":         nextID("ntf", len(store.data.Notifications)),
				"type":       "order_delivered",
				"title":      "Order delivered",
				"body":       "" + orderID + " delivered. Earnings updated.",
				"created_at": nowISO(),
				"read":       false,
			}}, store.data.Notifications...)
		}

		c.JSON(200, gin.H{"message": message, "order": order})
		return
	}

	c.JSON(404, gin.H{"error": "order_not_found"})
}

// AcceptOrder marks order as accepted.
func AcceptOrder(c *gin.Context) {
	updateOrderStatus(c, "accepted", "order_accepted")
}

// PickedUpOrder marks order as picked up.
func PickedUpOrder(c *gin.Context) {
	updateOrderStatus(c, "picked_up", "order_picked_up")
}

// DeliverOrder marks order as delivered.
func DeliverOrder(c *gin.Context) {
	updateOrderStatus(c, "delivered", "order_delivered")
}
