package worker

import (
	"fmt"
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

	if hasDB() {
		workerIDUint, parseWorkerErr := parseWorkerID(workerID)
		orderNumID, parseOrderErr := parseOrderID(orderID)
		if parseWorkerErr == nil && parseOrderErr == nil {
			type row struct {
				ID              uint    `gorm:"column:id"`
				Status          string  `gorm:"column:status"`
				PickupArea      string  `gorm:"column:pickup_area"`
				DropArea        string  `gorm:"column:drop_area"`
				DistanceKm      float64 `gorm:"column:distance_km"`
				TipInr          float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr  float64 `gorm:"column:delivery_fee_inr"`
				CreatedAt       string  `gorm:"column:created_at"`
				ZoneName        string  `gorm:"column:zone_name"`
				ZoneLevel       string  `gorm:"column:zone_level"`
				SourceNode      string  `gorm:"column:source_node"`
				DestinationNode string  `gorm:"column:destination_node"`
				CurrentNode     string  `gorm:"column:current_node"`
				Route           string  `gorm:"column:route"`
			}
			var r row
			err := workerDB.Raw(`
				SELECT 
					o.id,
					o.status,
					COALESCE(o.pickup_area, 'Pickup Location') AS pickup_area,
					COALESCE(o.drop_area, 'Drop Location') AS drop_area,
					COALESCE(o.distance_km, 0) AS distance_km,
					COALESCE(o.tip_inr, 0) AS tip_inr,
					COALESCE(o.delivery_fee_inr, 0) AS delivery_fee_inr,
					o.created_at::text,
					COALESCE(z.name, '') AS zone_name,
					COALESCE(z.level, '') AS zone_level,
					COALESCE(o.pickup_area, '') AS source_node,
					COALESCE(o.drop_area, '') AS destination_node,
					CASE 
						WHEN o.status = 'delivered' THEN COALESCE(o.drop_area, '')
						WHEN o.status = 'picked_up' THEN 'In transit'
						ELSE COALESCE(o.pickup_area, '')
					END AS current_node,
					COALESCE(o.pickup_area, '') || ' -> ' || COALESCE(o.drop_area, '') AS route
				FROM orders o
				LEFT JOIN zones z ON z.id = o.zone_id
				WHERE o.id = ? AND (o.worker_id = ? OR o.worker_id IS NULL OR o.status = 'assigned')
				LIMIT 1
			`, orderNumID, workerIDUint).Scan(&r).Error
			if err == nil && r.ID != 0 {
				c.JSON(200, gin.H{
					"order_id":         fmt.Sprintf("ord-%03d", r.ID),
					"pickup_area":      r.PickupArea,
					"drop_area":        r.DropArea,
					"distance_km":      r.DistanceKm,
					"earning_inr":      totalDeliveryEarningINR(r.TipInr),
					"tip_inr":          r.TipInr,
					"status":           r.Status,
					"assigned_at":      r.CreatedAt,
					"customer_name":    "Customer",
					"customer_phone":   "",
					"address":          "",
					"payment_type":     "cod",
					"zone_name":        r.ZoneName,
					"zone_level":       r.ZoneLevel,
					"source_node":      r.SourceNode,
					"destination_node": r.DestinationNode,
					"current_node":     r.CurrentNode,
					"route":            r.Route,
				})
				return
			}
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, order := range store.data.Orders {
		if order["order_id"] == orderID {
			c.JSON(200, order)
			return
		}
	}

	c.JSON(404, gin.H{"error": "order_not_found"})
}

func SendCustomerCode(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	orderID := c.Param("order_id")
	c.JSON(200, gin.H{
		"message":       "customer_code_sent",
		"order_id":      orderID,
		"customer_code": "1234",
	})
}

func SendFetchVerificationCode(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	c.JSON(200, gin.H{"message": "verification_code_sent", "code": "ZONE123"})
}

func VerifyFetchVerificationCode(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	body := parseBody(c)
	code := strings.TrimSpace(strings.ToUpper(bodyString(body, "code", "")))
	if code == "" {
		c.JSON(400, gin.H{"message": "verification_failed"})
		return
	}
	validCodes := map[string]bool{"ZONE123": true, "1234": true, "TAMBARAM": true}
	if !validCodes[code] {
		c.JSON(400, gin.H{"message": "verification_failed"})
		return
	}
	c.JSON(200, gin.H{"message": "verification_successful"})
}

func GetZoneConfig(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			type row struct {
				ZoneID   uint   `gorm:"column:zone_id"`
				ZoneName string `gorm:"column:zone_name"`
			}
			var r row
			_ = workerDB.Raw(`
				SELECT COALESCE(wp.zone_id, 0) AS zone_id, COALESCE(z.name, 'Tambaram') AS zone_name
				FROM worker_profiles wp
				LEFT JOIN zones z ON z.id = wp.zone_id
				WHERE wp.worker_id = ?
				LIMIT 1
			`, workerIDUint).Scan(&r).Error
			if r.ZoneName != "" {
				c.JSON(200, gin.H{
					"zone_id":               fmt.Sprintf("%d", r.ZoneID),
					"name":                  r.ZoneName,
					"require_ip_validation": false,
				})
				return
			}
		}
	}

	c.JSON(200, gin.H{
		"zone_id":               "1",
		"name":                  "Tambaram",
		"require_ip_validation": false,
	})
}

func GetSession(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	sessionID := c.Param("session_id")

	store.mu.RLock()
	defer store.mu.RUnlock()

	deliveriesCompleted := 0
	earningsInSession := 0
	for _, order := range store.data.Orders {
		if order["status"] == "delivered" {
			deliveriesCompleted++
			switch earning := order["earning_inr"].(type) {
			case int:
				earningsInSession += earning
			case float64:
				earningsInSession += int(earning)
			}
		}
	}

	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		if completed, ok := profile["orders_completed"].(int); ok && completed > deliveriesCompleted {
			deliveriesCompleted = completed
		}
	}

	c.JSON(200, gin.H{
		"session_id":           sessionID,
		"start_time":           timeNowMinus(90),
		"end_time":             nil,
		"status":               "active",
		"deliveries_completed": deliveriesCompleted,
		"earnings_in_session":  float64(earningsInSession),
	})
}

func GetSessionDeliveries(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	deliveries := make([]map[string]any, 0)
	for _, order := range store.data.Orders {
		status, _ := order["status"].(string)
		if status == "accepted" || status == "picked_up" || status == "delivered" {
			deliveries = append(deliveries, order)
		}
	}
	c.JSON(200, gin.H{"orders": deliveries})
}

func GetSessionFraudSignals(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	c.JSON(200, gin.H{
		"signals": []gin.H{
			{"type": "gps_consistency", "severity": "low", "timestamp": nowISO()},
			{"type": "idle_time", "severity": "low", "timestamp": nowISO()},
		},
	})
}

func EndSession(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	c.JSON(200, gin.H{"message": "session_ended", "session_id": c.Param("session_id")})
}

func DemoAssignOrders(c *gin.Context) {
	DemoSimulateOrders(c)
}

func DemoSimulateDeliveries(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	body := parseBody(c)
	targetCount := bodyInt(body, "count", 1)
	if targetCount <= 0 {
		targetCount = 1
	}

	processed := 0
	store.mu.Lock()
	for _, order := range store.data.Orders {
		if processed >= targetCount {
			break
		}
		status, _ := order["status"].(string)
		if status == "assigned" || status == "accepted" || status == "picked_up" {
			order["status"] = "delivered"
			order["updated_at"] = nowISO()
			processed++
		}
	}
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		if completed, ok := profile["orders_completed"].(int); ok {
			profile["orders_completed"] = completed + processed
		}
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "deliveries_simulated", "count": processed})
}

func timeNowMinus(minutes int) string {
	return time.Now().UTC().Add(-time.Duration(minutes) * time.Minute).Format(time.RFC3339)
}
