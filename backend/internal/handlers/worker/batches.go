package worker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type batchOrderRow struct {
	ID              uint    `gorm:"column:id"`
	ZoneID          uint    `gorm:"column:zone_id"`
	FromCity        string  `gorm:"column:from_city"`
	ToCity          string  `gorm:"column:to_city"`
	FromState       string  `gorm:"column:from_state"`
	ToState         string  `gorm:"column:to_state"`
	PickupArea      string  `gorm:"column:pickup_area"`
	DropArea        string  `gorm:"column:drop_area"`
	PackageWeightKg float64 `gorm:"column:package_weight_kg"`
	Status          string  `gorm:"column:status"`
	CreatedAt       string  `gorm:"column:created_at"`
}

type acceptBatchRequest struct {
	OrderIDs []string `json:"orderIds"`
}

func inferBatchZoneLevel(fromCity, toCity, fromState, toState string) string {
	fCity := strings.TrimSpace(strings.ToLower(fromCity))
	tCity := strings.TrimSpace(strings.ToLower(toCity))
	fState := strings.TrimSpace(strings.ToLower(fromState))
	tState := strings.TrimSpace(strings.ToLower(toState))

	if fCity != "" && fCity == tCity {
		return "A"
	}
	if fState != "" && fState == tState {
		return "B"
	}
	return "C"
}

func batchCodePart(value string, length int) string {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	if clean == "" {
		clean = "X"
	}
	if len(clean) >= length {
		return clean[:length]
	}
	return clean + strings.Repeat("X", length-len(clean))
}

func buildBatchID(zoneLevel, fromCity, toCity, fromState, toState, timestamp string) string {
	zone := strings.ToUpper(strings.TrimSpace(zoneLevel))
	if zone == "" {
		zone = "A"
	}
	cityCode := ""
	switch zone {
	case "A":
		cityCode = batchCodePart(fromCity, 6)
	default:
		cityCode = batchCodePart(fromCity, 3) + batchCodePart(toCity, 3)
	}

	stateCode := ""
	switch zone {
	case "C":
		stateCode = batchCodePart(fromState, 2) + batchCodePart(toState, 2)
	default:
		base := fromState
		if strings.TrimSpace(base) == "" {
			base = toState
		}
		stateCode = batchCodePart(base, 4)
	}

	return zone + cityCode + stateCode + timestamp
}

func normalizedWeight(weight float64) float64 {
	if weight > 0 {
		return weight
	}
	return 1.2
}

func batchStatusFromRows(rows []batchOrderRow, fallback string) string {
	status := strings.ToLower(strings.TrimSpace(fallback))
	for _, row := range rows {
		switch strings.ToLower(strings.TrimSpace(row.Status)) {
		case "picked_up":
			return "Picked Up"
		case "accepted":
			status = "accepted"
		case "assigned":
			if status == "" {
				status = "pending"
			}
		}
	}

	switch status {
	case "accepted":
		return "Accepted"
	case "pending":
		return "Pending"
	case "out for delivery":
		return "Out for Delivery"
	default:
		if status == "" {
			return "Pending"
		}
		if len(status) == 0 {
			return "Pending"
		}
		return strings.ToUpper(status[:1]) + status[1:]
	}
}

func rowsToBatches(rows []batchOrderRow, status string) []gin.H {
	type grouped struct {
		FromCity  string
		ToCity    string
		FromState string
		ToState   string
		ZoneLevel string
		Rows      []batchOrderRow
	}

	groups := map[string]*grouped{}
	for _, row := range rows {
		zoneLevel := inferBatchZoneLevel(row.FromCity, row.ToCity, row.FromState, row.ToState)
		key := fmt.Sprintf("%s|%s|%s|%s|%s", strings.ToLower(strings.TrimSpace(row.FromCity)), strings.ToLower(strings.TrimSpace(row.ToCity)), strings.ToLower(strings.TrimSpace(row.FromState)), strings.ToLower(strings.TrimSpace(row.ToState)), zoneLevel)
		if _, ok := groups[key]; !ok {
			groups[key] = &grouped{
				FromCity:  strings.TrimSpace(row.FromCity),
				ToCity:    strings.TrimSpace(row.ToCity),
				FromState: strings.TrimSpace(row.FromState),
				ToState:   strings.TrimSpace(row.ToState),
				ZoneLevel: zoneLevel,
				Rows:      []batchOrderRow{},
			}
		}
		groups[key].Rows = append(groups[key].Rows, row)
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	baseTime := time.Now().UTC()
	batches := make([]gin.H, 0, len(groups))
	for i, key := range keys {
		g := groups[key]
		timestamp := baseTime.Add(time.Duration(i) * time.Second).Format("20060102150405")
		batchID := buildBatchID(g.ZoneLevel, g.FromCity, g.ToCity, g.FromState, g.ToState, timestamp)
		orders := make([]gin.H, 0, len(g.Rows))
		totalWeight := 0.0
		for _, row := range g.Rows {
			weight := normalizedWeight(row.PackageWeightKg)
			totalWeight += weight
			orders = append(orders, gin.H{
				"orderId":         fmt.Sprintf("ord-%d", row.ID),
				"deliveryAddress": strings.TrimSpace(row.DropArea),
				"contactName":     "Customer",
				"contactPhone":    "N/A",
				"weight":          weight,
			})
		}

		fromCity := g.FromCity
		if fromCity == "" {
			fromCity = "Unknown"
		}
		toCity := g.ToCity
		if toCity == "" {
			toCity = fromCity
		}

		batches = append(batches, gin.H{
			"batchId":     batchID,
			"zoneLevel":   g.ZoneLevel,
			"fromCity":    fromCity,
			"toCity":      toCity,
			"totalWeight": totalWeight,
			"orderCount":  len(orders),
			"status":      batchStatusFromRows(g.Rows, status),
			"orders":      orders,
		})
	}

	return batches
}

func AcceptBatch(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	var req acceptBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.OrderIDs) == 0 {
		c.JSON(400, gin.H{"error": "invalid_batch_accept_request"})
		return
	}

	if !hasDB() {
		store.mu.Lock()
		defer store.mu.Unlock()

		now := time.Now().UTC().Format(time.RFC3339)
		accepted := make([]string, 0, len(req.OrderIDs))
		orderSet := make(map[string]struct{}, len(req.OrderIDs))
		for _, orderID := range req.OrderIDs {
			orderSet[orderID] = struct{}{}
		}

		for idx, order := range store.data.Orders {
			orderID := fmt.Sprintf("%v", order["order_id"])
			if _, ok := orderSet[orderID]; !ok {
				continue
			}
			order["worker_id"] = workerID
			order["status"] = "accepted"
			order["accepted_at"] = now
			order["updated_at"] = now
			store.data.Orders[idx] = order
			accepted = append(accepted, orderID)
		}

		if len(accepted) == 0 {
			c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
			return
		}

		c.JSON(200, gin.H{
			"message":          "batch_accepted",
			"batchId":          c.Param("batch_id"),
			"acceptedOrderIds": accepted,
		})
		return
	}

	workerIDUint, parseErr := parseWorkerID(workerID)
	if parseErr != nil {
		c.JSON(400, gin.H{"error": "invalid_worker_id"})
		return
	}

	accepted := make([]string, 0, len(req.OrderIDs))
	for _, orderID := range req.OrderIDs {
		orderNumID, parseOrderErr := parseOrderID(orderID)
		if parseOrderErr != nil {
			continue
		}

		result := workerDB.Exec(`
			UPDATE orders
			SET worker_id = ?, status = 'accepted', accepted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND (worker_id = ? OR status = 'assigned')
		`, workerIDUint, orderNumID, workerIDUint)
		if result.Error != nil || result.RowsAffected == 0 {
			continue
		}
		accepted = append(accepted, orderID)
	}

	if len(accepted) == 0 {
		c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
		return
	}

	c.JSON(200, gin.H{
		"message":          "batch_accepted",
		"batchId":          c.Param("batch_id"),
		"acceptedOrderIds": accepted,
	})
}

func GetAssignedBatches(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if !hasDB() {
		c.JSON(200, gin.H{"batches": []gin.H{}})
		return
	}

	workerIDUint, parseErr := parseWorkerID(workerID)
	if parseErr != nil {
		c.JSON(200, gin.H{"batches": []gin.H{}})
		return
	}

	rows := make([]batchOrderRow, 0)
	err := workerDB.Raw(`
		SELECT o.id, o.zone_id,
			COALESCE(o.from_city, '') AS from_city,
			COALESCE(o.to_city, '') AS to_city,
			COALESCE(o.from_state, '') AS from_state,
			COALESCE(o.to_state, '') AS to_state,
			COALESCE(o.pickup_area, '') AS pickup_area,
			COALESCE(o.drop_area, '') AS drop_area,
			COALESCE(o.package_weight_kg, 0) AS package_weight_kg,
			COALESCE(o.status, 'assigned') AS status,
			o.created_at::text AS created_at
		FROM orders o
		WHERE o.worker_id = ?
		  AND o.status IN ('assigned', 'accepted', 'picked_up')
		ORDER BY o.created_at DESC
	`, workerIDUint).Scan(&rows).Error
	if err != nil {
		c.JSON(200, gin.H{"batches": []gin.H{}})
		return
	}

	batches := rowsToBatches(rows, "Accepted")
	c.JSON(200, gin.H{"batches": batches})
}

func GetAvailableBatches(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	if !hasDB() {
		c.JSON(200, gin.H{"batches": []gin.H{}})
		return
	}

	rows := make([]batchOrderRow, 0)
	query := `
		SELECT o.id, o.zone_id,
			COALESCE(o.from_city, '') AS from_city,
			COALESCE(o.to_city, '') AS to_city,
			COALESCE(o.from_state, '') AS from_state,
			COALESCE(o.to_state, '') AS to_state,
			COALESCE(o.pickup_area, '') AS pickup_area,
			COALESCE(o.drop_area, '') AS drop_area,
			COALESCE(o.package_weight_kg, 0) AS package_weight_kg,
			COALESCE(o.status, 'assigned') AS status,
			o.created_at::text AS created_at
		FROM orders o
		WHERE o.status = 'assigned'
		ORDER BY o.created_at DESC
		LIMIT ?
	`

	err := workerDB.Raw(query, limit).Scan(&rows).Error
	if err != nil {
		c.JSON(200, gin.H{"batches": []gin.H{}})
		return
	}

	if workerID, ok := optionalAuthWorkerID(c); ok {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			scope := getWorkerOrderScope(workerIDUint)
			if scope.ZoneID != 0 {
				zoneNameLower := strings.ToLower(strings.TrimSpace(scope.ZoneName))
				zoneCityLower := strings.ToLower(strings.TrimSpace(scope.ZoneCity))
				filtered := make([]batchOrderRow, 0, len(rows))
				for _, row := range rows {
					fromCityLower := strings.ToLower(strings.TrimSpace(row.FromCity))
					pickupLower := strings.ToLower(strings.TrimSpace(row.PickupArea))
					if row.ZoneID == scope.ZoneID ||
						(zoneNameLower != "" && (fromCityLower == zoneNameLower || pickupLower == zoneNameLower)) ||
						(zoneCityLower != "" && (fromCityLower == zoneCityLower || pickupLower == zoneCityLower)) {
						filtered = append(filtered, row)
					}
				}
				rows = filtered
			}
		}
	}

	batches := rowsToBatches(rows, "Pending")
	c.JSON(200, gin.H{"batches": batches})
}
