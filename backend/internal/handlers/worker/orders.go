package worker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

func totalDeliveryEarningINR(tipINR float64) int {
	if tipINR < 0 {
		tipINR = 0
	}
	return deliveryBaseEarningINR + int(tipINR)
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

type workerOrderScope struct {
	ZoneID   uint
	ZoneName string
	ZoneCity string
}

func getWorkerOrderScope(workerIDUint uint) workerOrderScope {
	scope := workerOrderScope{}
	if !HasDB() {
		return scope
	}

	type scopeRow struct {
		ZoneID   uint   `gorm:"column:zone_id"`
		ZoneName string `gorm:"column:zone_name"`
		ZoneCity string `gorm:"column:zone_city"`
	}

	var row scopeRow
	err := workerDB.Raw(`
		SELECT wp.zone_id, COALESCE(z.name, '') AS zone_name, COALESCE(z.city, '') AS zone_city
		FROM worker_profiles wp
		LEFT JOIN zones z ON z.id = wp.zone_id
		WHERE wp.worker_id = ?
		LIMIT 1
	`, workerIDUint).Scan(&row).Error
	if err != nil {
		return scope
	}

	scope.ZoneID = row.ZoneID
	scope.ZoneName = strings.TrimSpace(row.ZoneName)
	scope.ZoneCity = strings.TrimSpace(row.ZoneCity)
	return scope
}

func optionalAuthWorkerID(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return "", false
	}

	if HasDB() {
		type tokenRow struct {
			UserID uint `gorm:"column:user_id"`
		}
		var row tokenRow
		err := workerDB.Raw(
			"SELECT user_id FROM auth_tokens WHERE token = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1",
			token,
		).Scan(&row).Error
		if err == nil && row.UserID != 0 {
			return fmt.Sprintf("%d", row.UserID), true
		}
	}

	if !allowInMemoryAuthFallback() {
		return "", false
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	workerID, ok := store.data.TokenToWorkerID[token]
	if !ok {
		return "", false
	}
	return workerID, true
}

// GetAssignedOrders returns assigned orders only.
func GetAssignedOrders(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	pathFilter := strings.TrimSpace(c.Query("path"))
	pathLike := "%" + strings.ToLower(pathFilter) + "%"

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type orderRow struct {
				ID             uint    `gorm:"column:id"`
				OrderValue     float64 `gorm:"column:order_value"`
				PackageSize    string  `gorm:"column:package_size"`
				PackageWeight  float64 `gorm:"column:package_weight_kg"`
				TipInr         float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
				ZoneRoutePath  string  `gorm:"column:zone_route_path"`
				FromCity       string  `gorm:"column:from_city"`
				ToCity         string  `gorm:"column:to_city"`
				FromState      string  `gorm:"column:from_state"`
				ToState        string  `gorm:"column:to_state"`
				Status         string  `gorm:"column:status"`
				PickupArea     string  `gorm:"column:pickup_area"`
				DropArea       string  `gorm:"column:drop_area"`
				DistanceKm     float64 `gorm:"column:distance_km"`
				CreatedAt      string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			baseSelect := "SELECT o.id, o.order_value, COALESCE(o.package_size, '') AS package_size, COALESCE(o.package_weight_kg, 0) AS package_weight_kg, COALESCE(o.tip_inr, 0) AS tip_inr, COALESCE(o.delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(o.zone_route_path, '[\"A\"]') AS zone_route_path, COALESCE(o.from_city, '') AS from_city, COALESCE(o.to_city, '') AS to_city, COALESCE(o.from_state, '') AS from_state, COALESCE(o.to_state, '') AS to_state, o.status, COALESCE(o.pickup_area, 'Pickup Location') AS pickup_area, COALESCE(o.drop_area, 'Drop Location') AS drop_area, COALESCE(o.distance_km, 0) AS distance_km, o.created_at::text FROM orders o"
			query := baseSelect + " WHERE o.worker_id = ?"
			args := []interface{}{workerIDUint}
			if pathFilter != "" {
				query += " AND (LOWER(COALESCE(o.pickup_area, '')) LIKE ? OR LOWER(COALESCE(o.drop_area, '')) LIKE ? OR LOWER(COALESCE(o.from_city, '')) LIKE ? OR LOWER(COALESCE(o.to_city, '')) LIKE ?)"
				args = append(args, pathLike, pathLike, pathLike, pathLike)
			}
			query += " ORDER BY o.created_at DESC LIMIT 20"
			_ = workerDB.Raw(query, args...).Scan(&rows).Error

			orders := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				zonePath := decodeZonePath(row.ZoneRoutePath)
				deliveryFee := row.DeliveryFeeInr
				if deliveryFee <= 0 {
					deliveryFee = float64(computeZoneRouteDeliveryFee(zonePath))
				}
				orders = append(orders, gin.H{
					"order_id":           fmt.Sprintf("ord-%03d", row.ID),
					"order_value":        row.OrderValue,
					"package_size":       row.PackageSize,
					"package_weight_kg":  row.PackageWeight,
					"from_city":          row.FromCity,
					"to_city":            row.ToCity,
					"from_state":         row.FromState,
					"to_state":           row.ToState,
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"earning_inr":        totalDeliveryEarningINR(row.TipInr),
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
		if worker, ok := order["worker_id"]; ok && fmt.Sprintf("%v", worker) != workerID {
			continue
		}
		if order["status"] == "assigned" || order["status"] == "accepted" || order["status"] == "picked_up" {
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
	pathFilter := strings.TrimSpace(c.Query("path"))
	pathLike := "%" + strings.ToLower(pathFilter) + "%"

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type orderRow struct {
				ID             uint    `gorm:"column:id"`
				OrderValue     float64 `gorm:"column:order_value"`
				PackageSize    string  `gorm:"column:package_size"`
				PackageWeight  float64 `gorm:"column:package_weight_kg"`
				TipInr         float64 `gorm:"column:tip_inr"`
				DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
				ZoneRoutePath  string  `gorm:"column:zone_route_path"`
				FromCity       string  `gorm:"column:from_city"`
				ToCity         string  `gorm:"column:to_city"`
				FromState      string  `gorm:"column:from_state"`
				ToState        string  `gorm:"column:to_state"`
				Status         string  `gorm:"column:status"`
				PickupArea     string  `gorm:"column:pickup_area"`
				DropArea       string  `gorm:"column:drop_area"`
				DistanceKm     float64 `gorm:"column:distance_km"`
				CreatedAt      string  `gorm:"column:created_at"`
			}
			rows := make([]orderRow, 0)
			query := "SELECT o.id, o.order_value, COALESCE(o.package_size, '') AS package_size, COALESCE(o.package_weight_kg, 0) AS package_weight_kg, COALESCE(o.tip_inr, 0) AS tip_inr, COALESCE(o.delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(o.zone_route_path, '[\"A\"]') AS zone_route_path, COALESCE(o.from_city, '') AS from_city, COALESCE(o.to_city, '') AS to_city, COALESCE(o.from_state, '') AS from_state, COALESCE(o.to_state, '') AS to_state, o.status, COALESCE(o.pickup_area, 'Pickup Location') AS pickup_area, COALESCE(o.drop_area, 'Drop Location') AS drop_area, COALESCE(o.distance_km, 0) AS distance_km, o.created_at::text FROM orders o WHERE o.worker_id = ?"
			args := []interface{}{workerIDUint}
			if pathFilter != "" {
				query += " AND (LOWER(COALESCE(o.pickup_area, '')) LIKE ? OR LOWER(COALESCE(o.drop_area, '')) LIKE ? OR LOWER(COALESCE(o.from_city, '')) LIKE ? OR LOWER(COALESCE(o.to_city, '')) LIKE ?)"
				args = append(args, pathLike, pathLike, pathLike, pathLike)
			}
			query += " ORDER BY o.created_at DESC LIMIT 50"
			_ = workerDB.Raw(query, args...).Scan(&rows).Error
			if len(rows) == 0 {
				// Demo fallback: return assigned pool so the app always has order cards to render.
				query = "SELECT o.id, o.order_value, COALESCE(o.package_size, '') AS package_size, COALESCE(o.package_weight_kg, 0) AS package_weight_kg, COALESCE(o.tip_inr, 0) AS tip_inr, COALESCE(o.delivery_fee_inr, 0) AS delivery_fee_inr, COALESCE(o.zone_route_path, '[\"A\"]') AS zone_route_path, COALESCE(o.from_city, '') AS from_city, COALESCE(o.to_city, '') AS to_city, COALESCE(o.from_state, '') AS from_state, COALESCE(o.to_state, '') AS to_state, o.status, COALESCE(o.pickup_area, 'Pickup Location') AS pickup_area, COALESCE(o.drop_area, 'Drop Location') AS drop_area, COALESCE(o.distance_km, 0) AS distance_km, o.created_at::text FROM orders o WHERE o.status = 'assigned'"
				args = []interface{}{}
				if pathFilter != "" {
					query += " AND (LOWER(COALESCE(o.pickup_area, '')) LIKE ? OR LOWER(COALESCE(o.drop_area, '')) LIKE ? OR LOWER(COALESCE(o.from_city, '')) LIKE ? OR LOWER(COALESCE(o.to_city, '')) LIKE ?)"
					args = append(args, pathLike, pathLike, pathLike, pathLike)
				}
				query += " ORDER BY o.created_at DESC LIMIT 50"
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
					"order_id":           fmt.Sprintf("ord-%03d", row.ID),
					"order_value":        row.OrderValue,
					"package_size":       row.PackageSize,
					"package_weight_kg":  row.PackageWeight,
					"pickup_area":        row.PickupArea,
					"drop_area":          row.DropArea,
					"distance_km":        row.DistanceKm,
					"earning_inr":        totalDeliveryEarningINR(row.TipInr),
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

	if HasDB() {
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
			}

			type row struct {
				ID         uint    `gorm:"column:id"`
				OrderValue float64 `gorm:"column:order_value"`
				TipInr     float64 `gorm:"column:tip_inr"`
				Status     string  `gorm:"column:status"`
				CreatedAt  string  `gorm:"column:created_at"`
				UpdatedAt  string  `gorm:"column:updated_at"`
			}
			var r row
			err := workerDB.Raw("SELECT id, order_value, COALESCE(tip_inr, 0) AS tip_inr, status, created_at::text, updated_at::text FROM orders WHERE id = ? AND worker_id = ?", orderNumID, workerIDUint).Scan(&r).Error
			if err == nil && r.ID != 0 {
				mirrorStoredOrderStatus(orderID, fmt.Sprintf("%d", workerIDUint), r.Status)
				refreshBatchSnapshotForOrderID(availableBatchCacheScope, orderID)
				refreshBatchSnapshotForOrderID(fmt.Sprintf("%d", workerIDUint), orderID)
				c.JSON(200, gin.H{"message": message, "order": gin.H{
					"order_id":    fmt.Sprintf("ord-%03d", r.ID),
					"pickup_area": "Tambaram",
					"drop_area":   "Camp Road",
					"distance_km": 3.1,
					"earning_inr": totalDeliveryEarningINR(r.TipInr),
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
	var responseOrder map[string]any

	for _, order := range store.data.Orders {
		if order["order_id"] != orderID {
			continue
		}

		previousStatus, _ := order["status"].(string)
		order["status"] = newStatus
		order["updated_at"] = nowISO()

		if newStatus == "delivered" && previousStatus != "delivered" {
			tip, _ := order["tip_inr"].(float64)
			earning := totalDeliveryEarningINR(tip)
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
				"body":       fmt.Sprintf("%s delivered. Earnings updated (+Rs %d incl. tip).", orderID, earning),
				"created_at": nowISO(),
				"read":       false,
			}}, store.data.Notifications...)
		}
		responseOrder = order
		break
	}
	store.mu.Unlock()

	if responseOrder != nil {
		refreshBatchSnapshotForOrder(availableBatchCacheScope, responseOrder)
		refreshBatchSnapshotForOrder(workerID, responseOrder)
		c.JSON(200, gin.H{"message": message, "order": responseOrder})
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
	pathFilter := strings.TrimSpace(c.Query("path"))
	pathLike := "%" + strings.ToLower(pathFilter) + "%"

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	if HasDB() {
		var workerScope workerOrderScope
		if workerID, ok := optionalAuthWorkerID(c); ok {
			if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
				workerScope = getWorkerOrderScope(workerIDUint)
			}
		}
		zoneNameLower := strings.ToLower(workerScope.ZoneName)
		zoneCityLower := strings.ToLower(workerScope.ZoneCity)

		type availableOrderRow struct {
			ID             uint    `gorm:"column:id"`
			OrderValue     float64 `gorm:"column:order_value"`
			PackageSize    string  `gorm:"column:package_size"`
			PackageWeight  float64 `gorm:"column:package_weight_kg"`
			TipInr         float64 `gorm:"column:tip_inr"`
			DeliveryFeeInr float64 `gorm:"column:delivery_fee_inr"`
			ZoneRoutePath  string  `gorm:"column:zone_route_path"`
			ZoneID         uint    `gorm:"column:zone_id"`
			ZoneName       string  `gorm:"column:zone_name"`
			Status         string  `gorm:"column:status"`
			FromCity       string  `gorm:"column:from_city"`
			ToCity         string  `gorm:"column:to_city"`
			FromState      string  `gorm:"column:from_state"`
			ToState        string  `gorm:"column:to_state"`
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
				COALESCE(o.package_size, '') as package_size,
				COALESCE(o.package_weight_kg, 0) as package_weight_kg,
				COALESCE(o.tip_inr, 0) as tip_inr,
				COALESCE(o.delivery_fee_inr, 0) as delivery_fee_inr,
				COALESCE(o.zone_route_path, '["A"]') as zone_route_path,
				o.zone_id, 
				z.name as zone_name, 
				o.status, 
				COALESCE(o.from_city, '') as from_city,
				COALESCE(o.to_city, '') as to_city,
				COALESCE(o.from_state, '') as from_state,
				COALESCE(o.to_state, '') as to_state,
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
		} else if workerScope.ZoneID != 0 {
			// Worker visibility rule: selected zone + cross-zone routes originating from worker's current location.
			if zoneNameLower != "" || zoneCityLower != "" {
				query += " AND (o.zone_id = ? OR LOWER(COALESCE(o.from_city, '')) = ? OR LOWER(COALESCE(o.pickup_area, '')) = ? OR LOWER(COALESCE(o.from_city, '')) = ? OR LOWER(COALESCE(o.pickup_area, '')) = ?)"
				args = append(args, workerScope.ZoneID, zoneNameLower, zoneNameLower, zoneCityLower, zoneCityLower)
			} else {
				query += " AND o.zone_id = ?"
				args = append(args, workerScope.ZoneID)
			}
		}
		if pathFilter != "" {
			query += " AND (LOWER(COALESCE(o.pickup_area, '')) LIKE ? OR LOWER(COALESCE(o.drop_area, '')) LIKE ? OR LOWER(COALESCE(o.from_city, '')) LIKE ? OR LOWER(COALESCE(o.to_city, '')) LIKE ?)"
			args = append(args, pathLike, pathLike, pathLike, pathLike)
		}

		query += " ORDER BY o.created_at DESC LIMIT ?"
		args = append(args, limit)

		err := workerDB.Raw(query, args...).Scan(&rows).Error
		if err == nil {
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
					"package_size":       row.PackageSize,
					"package_weight_kg":  row.PackageWeight,
					"from_city":          row.FromCity,
					"to_city":            row.ToCity,
					"from_state":         row.FromState,
					"to_state":           row.ToState,
					"earning_inr":        totalDeliveryEarningINR(row.TipInr),
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

	if HasDB() {
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
				COALESCE(o.tip_inr, 0) AS tip_inr,
				COALESCE(o.delivery_fee_inr, 0) AS delivery_fee_inr,
				COALESCE(o.zone_route_path, '["A"]') AS zone_route_path,
				o.worker_id,
				COALESCE(wp.name, 'Unknown') AS worker_name,
				o.zone_id, 
				z.name AS zone_name, 
				o.status, 
				COALESCE(o.pickup_area, 'Pickup') AS pickup_area,
				COALESCE(o.drop_area, 'Drop') AS drop_area,
				COALESCE(o.distance_km, 0) AS distance_km,
				o.created_at::text,
				o.delivered_at::text AS delivered_at
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
					"earning_inr":        totalDeliveryEarningINR(row.TipInr),
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
