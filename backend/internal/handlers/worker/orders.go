package worker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const deliveryBaseEarningINR = 60

var zoneBandDeliveryFeeINR = map[string]int{
	"A": 25,  // local / same city
	"B": 40,  // intra-state / regional
	"C": 65,  // metro-to-metro
	"D": 85,  // rest of India
	"E": 120, // northeast and J&K / difficult lanes
}

func parseOrderID(orderID string) (uint, error) {
	trimmed := strings.TrimSpace(orderID)
	trimmed = strings.TrimPrefix(trimmed, "ord-")
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func totalDeliveryEarningINR(feeINR, tipINR float64) int {
	if tipINR < 0 {
		tipINR = 0
	}
	if feeINR <= 0 {
		feeINR = float64(deliveryBaseEarningINR)
	}
	return int(feeINR) + int(tipINR)
}

func normalizeZoneBandPath(zonePath []string) []string {
	norm := make([]string, 0, len(zonePath))
	for _, z := range zonePath {
		band := strings.ToUpper(strings.TrimSpace(z))
		if _, ok := zoneBandDeliveryFeeINR[band]; ok {
			norm = append(norm, band)
		}
	}
	if len(norm) == 0 {
		return []string{"A"}
	}
	return norm
}

func computeZoneRouteDeliveryFee(zonePath []string) int {
	norm := normalizeZoneBandPath(zonePath)
	total := 0
	for _, band := range norm {
		total += zoneBandDeliveryFeeINR[band]
	}
	return total
}

func decodeZonePath(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"A"}
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		return normalizeZoneBandPath(arr)
	}
	if strings.Contains(raw, ">") {
		parts := strings.Split(raw, ">")
		return normalizeZoneBandPath(parts)
	}
	if strings.Contains(raw, ",") {
		parts := strings.Split(raw, ",")
		return normalizeZoneBandPath(parts)
	}
	return normalizeZoneBandPath([]string{raw})
}

func encodeZonePath(zonePath []string) string {
	norm := normalizeZoneBandPath(zonePath)
	b, err := json.Marshal(norm)
	if err != nil {
		return "[\"A\"]"
	}
	return string(b)
}

func zonePathDisplay(zonePath []string) string {
	norm := normalizeZoneBandPath(zonePath)
	return strings.Join(norm, ">")
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
				ID             uint    `gorm:"column:id"`
				OrderValue     float64 `gorm:"column:order_value"`
				TipInr         float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
				ZoneRoutePath  string  `gorm:"column:zone_route_path"`
				Status         string  `gorm:"column:status"`
				PickupArea     string  `gorm:"column:pickup_area"`
				DropArea       string  `gorm:"column:drop_area"`
				DistanceKm     float64 `gorm:"column:distance_km"`
				CreatedAt      string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			_ = workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, COALESCE(delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(zone_route_path, '[\"A\"]') AS zone_route_path, status, COALESCE(pickup_area, 'Velachery') AS pickup_area, COALESCE(drop_area, 'Adyar') AS drop_area, COALESCE(distance_km, 3.2) AS distance_km, created_at::text FROM orders WHERE worker_id = ? ORDER BY created_at DESC LIMIT 20", workerIDUint).Scan(&rows).Error
			if len(rows) == 0 {
				// Demo fallback: expose unclaimed assigned orders when worker has none.
				_ = workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, COALESCE(delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(zone_route_path, '[\"A\"]') AS zone_route_path, status, COALESCE(pickup_area, 'Chromepet') AS pickup_area, COALESCE(drop_area, 'Tambaram') AS drop_area, COALESCE(distance_km, 2.5) AS distance_km, created_at::text FROM orders WHERE status = 'assigned' ORDER BY created_at DESC LIMIT 20").Scan(&rows).Error
			}
			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				zonePath := decodeZonePath(row.ZoneRoutePath)
				deliveryFee := row.DeliveryFeeInr
				if deliveryFee <= 0 {
					deliveryFee = float64(computeZoneRouteDeliveryFee(zonePath))
				}
				orders = append(orders, gin.H{
					"order_id":           fmt.Sprintf("ord-%03d", row.ID),
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"earning_inr":        totalDeliveryEarningINR(deliveryFee, row.TipInr),
					"tip_inr":            row.TipInr,
					"delivery_fee_inr":   deliveryFee,
					"zone_route_path":    zonePath,
					"zone_route_display": zonePathDisplay(zonePath),
					"status":             row.Status,
					"assigned_at":        row.CreatedAt,
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
				ID             uint    `gorm:"column:id"`
				OrderValue     float64 `gorm:"column:order_value"`
				TipInr         float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
				ZoneRoutePath  string  `gorm:"column:zone_route_path"`
				Status         string  `gorm:"column:status"`
				PickupArea     string  `gorm:"column:pickup_area"`
				DropArea       string  `gorm:"column:drop_area"`
				DistanceKm     float64 `gorm:"column:distance_km"`
				CreatedAt      string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			_ = workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, COALESCE(delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(zone_route_path, '[\"A\"]') AS zone_route_path, status, COALESCE(pickup_area, 'Selaiyur') AS pickup_area, COALESCE(drop_area, 'Sholinganallur') AS drop_area, COALESCE(distance_km, 5.1) AS distance_km, created_at::text FROM orders WHERE worker_id = ? ORDER BY created_at DESC LIMIT 50", workerIDUint).Scan(&rows).Error
			if len(rows) < 5 { // If worker has few orders, always show a pool of available ones
				var extraRows []orderRow
				_ = workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, COALESCE(delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(zone_route_path, '[\"A\"]') AS zone_route_path, status, COALESCE(pickup_area, 'Guindy') AS pickup_area, COALESCE(drop_area, 'T.Nagar') AS drop_area, COALESCE(distance_km, 4.3) AS distance_km, created_at::text FROM orders WHERE status = 'assigned' ORDER BY created_at DESC LIMIT 10").Scan(&extraRows).Error
				rows = append(rows, extraRows...)
			}
			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				zonePath := decodeZonePath(row.ZoneRoutePath)
				deliveryFee := row.DeliveryFeeInr
				if deliveryFee <= 0 {
					deliveryFee = float64(computeZoneRouteDeliveryFee(zonePath))
				}
				orders = append(orders, gin.H{
					"order_id":           fmt.Sprintf("ord-%03d", row.ID),
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"earning_inr":        totalDeliveryEarningINR(deliveryFee, row.TipInr),
					"tip_inr":            row.TipInr,
					"delivery_fee_inr":   deliveryFee,
					"zone_route_path":    zonePath,
					"zone_route_display": zonePathDisplay(zonePath),
					"status":             row.Status,
					"assigned_at":        row.CreatedAt,
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
				_ = workerDB.Exec("UPDATE orders SET worker_id = ?, status='accepted', accepted_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND (worker_id = ? OR status = 'assigned')", workerIDUint, orderNumID, workerIDUint).Error
			}
			if newStatus == "picked_up" {
				_ = workerDB.Exec("UPDATE orders SET status='picked_up', picked_up_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Error
			}
			if newStatus == "delivered" {
				_ = workerDB.Exec("UPDATE orders SET status='delivered', delivered_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Error
				_ = workerDB.Exec("INSERT INTO notifications (worker_id, type, message) VALUES (?, 'order_delivered', ?)", workerIDUint, fmt.Sprintf("%s delivered. Earnings updated.", orderID)).Error

				// Update earnings in DB
				var tip, fee float64
				_ = workerDB.Raw("SELECT COALESCE(tip_inr, 0), COALESCE(delivery_fee_inr, 0) FROM orders WHERE id = ?", orderNumID).Row().Scan(&tip, &fee)
				earning := totalDeliveryEarningINR(fee, tip)

				weekStart, _ := weekBounds(time.Now().UTC())

				// Update weekly summary
				_ = workerDB.Exec(`
					INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings, claim_eligible)
					VALUES (?, ?, ?, ?, false)
					ON CONFLICT (worker_id, week_start)
					DO UPDATE SET total_earnings = weekly_earnings_summary.total_earnings + ?, updated_at = CURRENT_TIMESTAMP
				`, workerIDUint, weekStart, weekStart.AddDate(0, 0, 7), float64(earning), float64(earning)).Error

				// Update daily record
				_ = workerDB.Exec(`
					INSERT INTO earnings_records (worker_id, date, amount_earned, hours_worked, created_at)
					VALUES (?, CURRENT_DATE, ?, 0, CURRENT_TIMESTAMP)
					ON CONFLICT (worker_id, date)
					DO UPDATE SET amount_earned = earnings_records.amount_earned + ?
				`, workerIDUint, float64(earning), float64(earning)).Error

				// Update lifetime earnings
				_ = workerDB.Exec("UPDATE worker_profiles SET total_earnings_lifetime = total_earnings_lifetime + ?, updated_at = CURRENT_TIMESTAMP WHERE worker_id = ?", float64(earning), workerIDUint).Error
			}

			type row struct {
				ID             uint    `gorm:"column:id"`
				OrderValue     float64 `gorm:"column:order_value"`
				TipInr         float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
				Status         string  `gorm:"column:status"`
				PickupArea     string  `gorm:"column:pickup_area"`
				DropArea       string  `gorm:"column:drop_area"`
				DistanceKm     float64 `gorm:"column:distance_km"`
				CreatedAt      string  `gorm:"column:created_at"`
				UpdatedAt      string  `gorm:"column:updated_at"`
			}
			var r row
			err := workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, COALESCE(delivery_fee_inr, 0) AS delivery_fee_inr, status, COALESCE(pickup_area, '') AS pickup_area, COALESCE(drop_area, '') AS drop_area, COALESCE(distance_km, 0) AS distance_km, created_at::text, updated_at::text FROM orders WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Scan(&r).Error
			if err == nil && r.ID != 0 {
				c.JSON(200, gin.H{"message": message, "order": gin.H{
					"order_id":    fmt.Sprintf("ord-%03d", r.ID),
					"pickup_area": r.PickupArea,
					"drop_area":   r.DropArea,
					"distance_km": r.DistanceKm,
					"earning_inr": totalDeliveryEarningINR(r.DeliveryFeeInr, r.TipInr),
					"tip_inr":     r.TipInr,
					"status":      r.Status,
					"assigned_at": r.CreatedAt,
					"updated_at":  r.UpdatedAt,
				}})
				return
			}

			c.JSON(404, gin.H{"error": "order_not_found_or_not_assignable"})
			return
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
			tip, _ := order["tip_inr"].(float64)
			fee, _ := order["delivery_fee_inr"].(float64)
			earning := totalDeliveryEarningINR(fee, tip)
			if current, ok := store.data.Earnings["this_week_actual"].(int); ok {
				store.data.Earnings["this_week_actual"] = current + earning
			} else if currentF, ok := store.data.Earnings["this_week_actual"].(float64); ok {
				store.data.Earnings["this_week_actual"] = int(currentF) + earning
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
				"body":       fmt.Sprintf("%s delivered. Earnings updated (+Rs %d incl. tip).", orderID, earning),
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

// GetAvailableOrders returns unassigned orders with zone information.
// Can filter by zone_id via query parameter.
// Used by fake_order_publisher to fetch orders from backend API.
func GetAvailableOrders(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	zoneIDStr := c.Query("zone_id")

	limit := 4
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	if hasDB() {
		type availableOrderRow struct {
			ID             uint    `gorm:"column:id"`
			OrderValue     float64 `gorm:"column:order_value"`
			TipInr         float64 `gorm:"column:tip_inr"`
			DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
			ZoneRoutePath  string  `gorm:"column:zone_route_path"`
			ZoneID         uint    `gorm:"column:zone_id"`
			ZoneName       string  `gorm:"column:zone_name"`
			Status         string  `gorm:"column:status"`
			PickupArea     string  `gorm:"column:pickup_area"`
			DropArea       string  `gorm:"column:drop_area"`
			DistanceKm     float64 `gorm:"column:distance_km"`
			CreatedAt      string  `gorm:"column:created_at"`
		}

		rows := make([]availableOrderRow, 0)
		query := `
			SELECT 
				o.id, 
				o.order_value, 
				COALESCE(o.tip_inr, 0) as tip_inr,
				COALESCE(o.delivery_fee_inr, 0) as delivery_fee_inr,
				COALESCE(o.zone_route_path, '["A"]') as zone_route_path,
				o.zone_id, 
				z.name as zone_name, 
				o.status, 
				COALESCE(o.pickup_area, 'Pickup Location') as pickup_area,
				COALESCE(o.drop_area, 'Drop Location') as drop_area,
				COALESCE(o.distance_km, 0) as distance_km,
				o.created_at::text
			FROM orders o
			LEFT JOIN zones z ON o.zone_id = z.id
			WHERE o.status = 'assigned'
		`

		args := []interface{}{}
		if zoneIDStr != "" {
			query += " AND o.zone_id = ?"
			args = append(args, zoneIDStr)
		}

		query += " ORDER BY o.created_at DESC LIMIT ?"
		args = append(args, limit)

		err := workerDB.Raw(query, args...).Scan(&rows).Error
		if err == nil {
			if len(rows) == 0 {
				zoneToUse := 1
				if zoneIDStr != "" {
					z, _ := strconv.Atoi(zoneIDStr)
					zoneToUse = z
				}
				var defaultUser int
				_ = workerDB.Raw("SELECT id FROM users LIMIT 1").Scan(&defaultUser)
				if defaultUser == 0 {
					defaultUser = 1
				}
				pickupOptions := []string{"Anna Nagar", "Mylapore", "Nungambakkam", "Besant Nagar", "Egmore"}
				dropOptions := []string{"Selaiyur", "Medavakkam", "Perungudi", "Kottivakkam", "Chemmancherry"}
				for i := 0; i < 5; i++ {
					p := pickupOptions[i%len(pickupOptions)]
					d := dropOptions[i%len(dropOptions)]
					_ = workerDB.Exec("INSERT INTO orders (worker_id, order_value, tip_inr, delivery_fee_inr, zone_id, status, pickup_area, drop_area, distance_km) VALUES (?, ?, ?, ?, ?, 'assigned', ?, ?, ?)", defaultUser, 50+i*10, 0, 30, zoneToUse, p, d, 2.5+float64(i)*0.5).Error
				}
				_ = workerDB.Raw(query, args...).Scan(&rows).Error
			}

			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				zonePath := decodeZonePath(row.ZoneRoutePath)
				deliveryFee := row.DeliveryFeeInr
				if deliveryFee <= 0 {
					deliveryFee = float64(computeZoneRouteDeliveryFee(zonePath))
				}
				orders = append(orders, gin.H{
					"order_id":           fmt.Sprintf("ord-%d", row.ID),
					"order_value":        row.OrderValue,
					"earning_inr":        totalDeliveryEarningINR(deliveryFee, row.TipInr),
					"tip_inr":            row.TipInr,
					"delivery_fee_inr":   deliveryFee,
					"zone_route_path":    zonePath,
					"zone_route_display": zonePathDisplay(zonePath),
					"zone_id":            row.ZoneID,
					"zone_name":          row.ZoneName,
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"status":             row.Status,
					"created_at":         row.CreatedAt,
				})
			}

			c.JSON(200, gin.H{
				"count":  len(orders),
				"orders": orders,
			})
			return
		}
	}

	// Fallback to in-memory store
	store.mu.RLock()
	defer store.mu.RUnlock()

	available := make([]map[string]any, 0)
	for _, order := range store.data.Orders {
		if order["status"] == "assigned" {
			available = append(available, order)
		}
		if len(available) >= limit {
			break
		}
	}

	c.JSON(200, gin.H{
		"count":  len(available),
		"orders": available,
	})
}

// GetDeliveries returns all delivered orders for deliveries tracking.
// Can filter by worker_id, zone_id, or date range via query parameters.
func GetDeliveries(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	workerIDStr := c.Query("worker_id")
	zoneIDStr := c.Query("zone_id")
	statusStr := c.DefaultQuery("status", "delivered")

	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	if hasDB() {
		type deliveryRow struct {
			ID             uint    `gorm:"column:id"`
			OrderValue     float64 `gorm:"column:order_value"`
			TipInr         float64 `gorm:"column:tip_inr"`
			DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
			ZoneRoutePath  string  `gorm:"column:zone_route_path"`
			WorkerID       uint    `gorm:"column:worker_id"`
			WorkerName     string  `gorm:"column:worker_name"`
			ZoneID         uint    `gorm:"column:zone_id"`
			ZoneName       string  `gorm:"column:zone_name"`
			Status         string  `gorm:"column:status"`
			PickupArea     string  `gorm:"column:pickup_area"`
			DropArea       string  `gorm:"column:drop_area"`
			DistanceKm     float64 `gorm:"column:distance_km"`
			CreatedAt      string  `gorm:"column:created_at"`
			DeliveredAt    *string `gorm:"column:delivered_at"`
		}

		rows := make([]deliveryRow, 0)
		query := `
			SELECT 
				o.id, 
				o.order_value, 
				COALESCE(o.tip_inr, 0) as tip_inr,
				COALESCE(o.delivery_fee_inr, 0) as delivery_fee_inr,
				COALESCE(o.zone_route_path, '["A"]') as zone_route_path,
				o.worker_id,
				COALESCE(wp.name, 'Unknown') as worker_name,
				o.zone_id, 
				z.name as zone_name, 
				o.status, 
				COALESCE(o.pickup_area, 'Pickup') as pickup_area,
				COALESCE(o.drop_area, 'Drop') as drop_area,
				COALESCE(o.distance_km, 0) as distance_km,
				o.created_at::text,
				o.delivered_at::text as delivered_at
			FROM orders o
			LEFT JOIN zones z ON o.zone_id = z.id
			LEFT JOIN worker_profiles wp ON o.worker_id = wp.worker_id
			WHERE o.status = ?
		`

		args := []interface{}{statusStr}

		if workerIDStr != "" {
			query += " AND o.worker_id = ?"
			args = append(args, workerIDStr)
		}
		if zoneIDStr != "" {
			query += " AND o.zone_id = ?"
			args = append(args, zoneIDStr)
		}

		query += " ORDER BY o.delivered_at DESC NULLS LAST, o.created_at DESC LIMIT ?"
		args = append(args, limit)

		err := workerDB.Raw(query, args...).Scan(&rows).Error
		if err == nil {
			deliveries := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				zonePath := decodeZonePath(row.ZoneRoutePath)
				deliveryFee := row.DeliveryFeeInr
				if deliveryFee <= 0 {
					deliveryFee = float64(computeZoneRouteDeliveryFee(zonePath))
				}
				delivery := gin.H{
					"order_id":           fmt.Sprintf("ord-%d", row.ID),
					"order_value":        row.OrderValue,
					"earning_inr":        totalDeliveryEarningINR(deliveryFee, row.TipInr),
					"tip_inr":            row.TipInr,
					"delivery_fee_inr":   deliveryFee,
					"zone_route_path":    zonePath,
					"zone_route_display": zonePathDisplay(zonePath),
					"worker_id":          row.WorkerID,
					"worker_name":        row.WorkerName,
					"zone_id":            row.ZoneID,
					"zone_name":          row.ZoneName,
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"status":             row.Status,
					"created_at":         row.CreatedAt,
				}
				if row.DeliveredAt != nil {
					delivery["delivered_at"] = *row.DeliveredAt
				}
				deliveries = append(deliveries, delivery)
			}

			c.JSON(200, gin.H{
				"count":      len(deliveries),
				"deliveries": deliveries,
			})
			return
		}
	}

	// Fallback to in-memory store
	store.mu.RLock()
	defer store.mu.RUnlock()

	delivered := make([]map[string]any, 0)
	for _, order := range store.data.Orders {
		if order["status"] == "delivered" {
			delivered = append(delivered, order)
		}
		if len(delivered) >= limit {
			break
		}
	}

	c.JSON(200, gin.H{
		"count":      len(delivered),
		"deliveries": delivered,
	})
}

func weekBounds(now time.Time) (time.Time, time.Time) {
	normalized := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(normalized.Weekday()) + 6) % 7
	start := normalized.AddDate(0, 0, -offset)
	end := start.AddDate(0, 0, 6)
	return start, end
}
