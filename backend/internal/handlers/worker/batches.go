package worker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	TipInr          float64 `gorm:"column:tip_inr"`
	OrderValue      float64 `gorm:"column:order_value"`
	CustomerName    string  `gorm:"column:customer_name"`
	CustomerPhone   string  `gorm:"column:customer_contact_number"`
	Address         string  `gorm:"column:address"`
	Status          string  `gorm:"column:status"`
	CreatedAt       string  `gorm:"column:created_at"`
	PickedUpAt      string  `gorm:"column:picked_up_at"`
	DeliveredAt     string  `gorm:"column:delivered_at"`
	WorkerID        uint    `gorm:"column:worker_id"`
}

type acceptBatchRequest struct {
	OrderIDs   []string `json:"orderIds"`
	PickupCode string   `json:"pickupCode"`
}

const (
	minOrderWeightKg    = 0.05
	maxOrderWeightKg    = 5.0
	maxBatchWeightKg    = 12.0
	targetBatchWeightKg = 10.0
)

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

func buildBatchID(zoneLevel, fromCity, toCity, fromState, toState string, timestamp time.Time) string {
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

	datetimeStr := timestamp.Format("20060102150405")
	return zone + cityCode + stateCode + datetimeStr
}

func batchTimestampForRows(rows []batchOrderRow) time.Time {
	if len(rows) == 0 {
		return time.Now().UTC()
	}

	var best time.Time
	for _, row := range rows {
		parsed, err := parseBatchTime(row.CreatedAt)
		if err != nil {
			continue
		}
		if best.IsZero() || parsed.Before(best) {
			best = parsed
		}
	}

	if best.IsZero() {
		return time.Now().UTC()
	}
	return best.UTC()
}

func parseBatchTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty batch time")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05.999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse batch time %q", value)
}

func pickupCodeFromBatchID(batchID string) string {
	seed := 0
	for _, r := range strings.ToUpper(strings.TrimSpace(batchID)) {
		seed = (seed*31 + int(r)) % 9000
	}
	return fmt.Sprintf("%04d", 1000+seed)
}

func deliveryCodeFromBatchID(batchID string) string {
	seed := 7
	for _, r := range strings.ToUpper(strings.TrimSpace(batchID)) {
		seed = (seed*37 + int(r)) % 9000
	}
	return fmt.Sprintf("%04d", 1000+seed)
}

func deliveryCodeFromOrderID(orderID string) string {
	seed := 11
	for _, r := range strings.ToUpper(strings.TrimSpace(orderID)) {
		seed = (seed*41 + int(r)) % 9000
	}
	return fmt.Sprintf("%04d", 1000+seed)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizedWeight(weight float64) float64 {
	if weight <= 0 {
		weight = 1.2
	}
	if weight < minOrderWeightKg {
		return minOrderWeightKg
	}
	if weight > maxOrderWeightKg {
		return maxOrderWeightKg
	}
	return weight
}

func batchStatusFromRows(rows []batchOrderRow, fallback string) string {
	status := strings.ToLower(strings.TrimSpace(fallback))
	hasPickedUp := false
	hasDelivered := false
	for _, row := range rows {
		switch strings.ToLower(strings.TrimSpace(row.Status)) {
		case "delivered":
			hasDelivered = true
		case "picked_up":
			hasPickedUp = true
		case "accepted":
			if status == "" {
				status = "assigned"
			}
		case "assigned":
			if status == "" {
				status = "assigned"
			}
		}
	}
	if hasDelivered {
		return "Delivered"
	}
	if hasPickedUp {
		return "Picked Up"
	}

	switch status {
	case "assigned":
		return "Assigned"
	default:
		if status == "" {
			return "Assigned"
		}
		return strings.ToUpper(status[:1]) + status[1:]
	}
}

func packRowsIntoBatches(rows []batchOrderRow) [][]batchOrderRow {
	if len(rows) == 0 {
		return nil
	}

	sortedRows := append([]batchOrderRow(nil), rows...)
	sort.SliceStable(sortedRows, func(i, j int) bool {
		left := normalizedWeight(sortedRows[i].PackageWeightKg)
		right := normalizedWeight(sortedRows[j].PackageWeightKg)
		if left == right {
			return sortedRows[i].CreatedAt < sortedRows[j].CreatedAt
		}
		return left > right
	})

	type packedBatch struct {
		rows        []batchOrderRow
		totalWeight float64
	}

	packed := make([]packedBatch, 0)
	for _, row := range sortedRows {
		weight := normalizedWeight(row.PackageWeightKg)
		bestIndex := -1
		bestScore := 0.0
		for index := range packed {
			projected := packed[index].totalWeight + weight
			if projected > maxBatchWeightKg {
				continue
			}
			score := targetBatchWeightKg - projected
			if score < 0 {
				score = -score
			}
			if packed[index].totalWeight < targetBatchWeightKg && projected >= targetBatchWeightKg {
				score -= 0.75
			}
			if bestIndex == -1 || score < bestScore {
				bestIndex = index
				bestScore = score
			}
		}
		if bestIndex == -1 {
			packed = append(packed, packedBatch{rows: []batchOrderRow{row}, totalWeight: weight})
			continue
		}
		packed[bestIndex].rows = append(packed[bestIndex].rows, row)
		packed[bestIndex].totalWeight += weight
	}

	result := make([][]batchOrderRow, 0, len(packed))
	for _, batch := range packed {
		result = append(result, batch.rows)
	}
	return result
}

func rowsToBatches(rows []batchOrderRow, status string, includeCodes bool) []gin.H {
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

	batches := make([]gin.H, 0, len(groups))
	for _, key := range keys {
		g := groups[key]
		packedRows := packRowsIntoBatches(g.Rows)
		batchTimestamp := batchTimestampForRows(g.Rows)
		for batchIndex, batchRows := range packedRows {
			batchID := fmt.Sprintf("%s-%02d", buildBatchID(g.ZoneLevel, g.FromCity, g.ToCity, g.FromState, g.ToState, batchTimestamp), batchIndex+1)
			orders := make([]gin.H, 0, len(batchRows))
			totalWeight := 0.0
			totalEarning := 0.0
			var pickupTime string
			var deliveryTime string
			for _, row := range batchRows {
				weight := normalizedWeight(row.PackageWeightKg)
				totalWeight += weight
				totalEarning += float64(totalDeliveryEarningINR(row.TipInr))
				orderID := fmt.Sprintf("ord-%d", row.ID)
				if pickupTime == "" {
					pickupTime = row.PickedUpAt
				}
				if deliveryTime == "" {
					deliveryTime = row.DeliveredAt
				}
				orderPayload := gin.H{
					"orderId":         orderID,
					"deliveryAddress": strings.TrimSpace(firstNonEmpty(row.Address, row.DropArea)),
					"contactName":     firstNonEmpty(row.CustomerName, "Customer"),
					"contactPhone":    firstNonEmpty(row.CustomerPhone, "N/A"),
					"weight":          weight,
					"pickupArea":      strings.TrimSpace(row.PickupArea),
					"dropArea":        strings.TrimSpace(row.DropArea),
					"status":          strings.TrimSpace(row.Status),
					"pickupTime":      row.PickedUpAt,
					"deliveryTime":    row.DeliveredAt,
				}
				if includeCodes && strings.EqualFold(strings.TrimSpace(g.ZoneLevel), "A") {
					orderPayload["deliveryCode"] = deliveryCodeFromOrderID(orderID)
				}
				orders = append(orders, orderPayload)
			}

			fromCity := g.FromCity
			if fromCity == "" {
				fromCity = "Unknown"
			}
			toCity := g.ToCity
			if toCity == "" {
				toCity = fromCity
			}

			batchPayload := gin.H{
				"batchId":         batchID,
				"batchKey":        fmt.Sprintf("%s#%02d", key, batchIndex+1),
				"batchGroupKey":   key,
				"batchIndex":      batchIndex + 1,
				"zoneLevel":       g.ZoneLevel,
				"fromCity":        fromCity,
				"toCity":          toCity,
				"totalWeight":     totalWeight,
				"targetWeight":    targetBatchWeightKg,
				"maxWeight":       maxBatchWeightKg,
				"orderCount":      len(orders),
				"status":          batchStatusFromRows(batchRows, status),
				"pickupTime":      pickupTime,
				"deliveryTime":    deliveryTime,
				"batchEarningInr": totalEarning,
				"orders":          orders,
			}
			if includeCodes {
				batchPayload["pickupCode"] = pickupCodeFromBatchID(batchID)
				batchPayload["deliveryCode"] = deliveryCodeFromBatchID(batchID)
			}

			batches = append(batches, batchPayload)
		}
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

	expectedCode := pickupCodeFromBatchID(c.Param("batch_id"))
	if strings.TrimSpace(req.PickupCode) == "" || strings.TrimSpace(req.PickupCode) != expectedCode {
		c.JSON(400, gin.H{"error": "incorrect_pickup_code"})
		return
	}

	if !HasDB() {
		store.mu.Lock()

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
			order["status"] = "picked_up"
			order["accepted_at"] = now
			order["picked_up_at"] = now
			order["updated_at"] = now
			store.data.Orders[idx] = order
			accepted = append(accepted, orderID)
		}

		if len(accepted) == 0 {
			store.mu.Unlock()
			c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
			return
		}

		store.mu.Unlock()

		if len(accepted) > 0 {
			refreshBatchSnapshotForOrderID(availableBatchCacheScope, accepted[0])
			refreshBatchSnapshotForOrderID(workerID, accepted[0])
		}

		c.JSON(200, gin.H{
			"message":          "batch_picked_up",
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

	batchID := strings.TrimSpace(c.Param("batch_id"))
	if batchID == "" {
		c.JSON(400, gin.H{"error": "invalid_batch_id"})
		return
	}

	if strings.TrimSpace(req.PickupCode) != pickupCodeFromBatchID(batchID) {
		c.JSON(400, gin.H{"error": "incorrect_pickup_code"})
		return
	}

	accepted := make([]string, 0, len(req.OrderIDs))
	err := workerDB.Transaction(func(tx *gorm.DB) error {
		orderNums := make([]uint, 0, len(req.OrderIDs))
		orderIDsByNum := make(map[uint]string, len(req.OrderIDs))
		for _, orderID := range req.OrderIDs {
			orderNumID, parseOrderErr := parseOrderID(orderID)
			if parseOrderErr != nil {
				continue
			}
			orderNums = append(orderNums, orderNumID)
			orderIDsByNum[orderNumID] = orderID
		}
		if len(orderNums) == 0 {
			return fmt.Errorf("batch_not_found_or_not_assignable")
		}

		var rows []batchOrderRow
		if err := tx.Raw(`
			SELECT o.id, o.zone_id,
				COALESCE(o.from_city, '') AS from_city,
				COALESCE(o.to_city, '') AS to_city,
				COALESCE(o.from_state, '') AS from_state,
				COALESCE(o.to_state, '') AS to_state,
				COALESCE(o.pickup_area, '') AS pickup_area,
				COALESCE(o.drop_area, '') AS drop_area,
				COALESCE(o.package_weight_kg, 0) AS package_weight_kg,
				COALESCE(o.tip_inr, 0) AS tip_inr,
				COALESCE(o.order_value, 0) AS order_value,
				COALESCE(o.customer_name, 'Customer') AS customer_name,
				COALESCE(o.customer_contact_number, 'N/A') AS customer_contact_number,
				COALESCE(o.address, COALESCE(o.drop_area, '')) AS address,
				COALESCE(o.status, 'assigned') AS status,
				o.created_at::text AS created_at
			FROM orders o
			WHERE o.id IN ?
			  AND (o.worker_id = ? OR o.status = 'assigned')
			ORDER BY o.created_at ASC
		`, orderNums, workerIDUint).Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(orderNums) {
			return fmt.Errorf("batch_not_found_or_not_assignable")
		}

		now := time.Now().UTC()
		nowText := now.Format(time.RFC3339)
		batchEarn := 0.0
		batchWeight := 0.0
		for _, row := range rows {
			batchEarn += float64(totalDeliveryEarningINR(row.TipInr))
			batchWeight += normalizedWeight(row.PackageWeightKg)
		}

		if err := tx.Exec(`
			INSERT INTO batches (batch_id, zone_level, from_city, to_city, total_weight, order_count, status, pickup_code, delivery_code, pickup_user_id, pickup_time, delivery_time, batch_earning_inr, earnings_posted, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'Picked Up', ?, ?, ?, ?, NULL, ?, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (batch_id) DO UPDATE SET
				zone_level = EXCLUDED.zone_level,
				from_city = EXCLUDED.from_city,
				to_city = EXCLUDED.to_city,
				total_weight = EXCLUDED.total_weight,
				order_count = EXCLUDED.order_count,
				status = EXCLUDED.status,
				pickup_code = EXCLUDED.pickup_code,
				delivery_code = EXCLUDED.delivery_code,
				pickup_user_id = EXCLUDED.pickup_user_id,
				pickup_time = EXCLUDED.pickup_time,
				batch_earning_inr = EXCLUDED.batch_earning_inr,
				updated_at = CURRENT_TIMESTAMP
		`, batchID, inferBatchZoneLevel(rows[0].FromCity, rows[0].ToCity, rows[0].FromState, rows[0].ToState), firstNonEmpty(rows[0].FromCity, "Unknown"), firstNonEmpty(rows[0].ToCity, rows[0].FromCity), batchWeight, len(rows), pickupCodeFromBatchID(batchID), deliveryCodeFromBatchID(batchID), workerIDUint, nowText, batchEarn).Error; err != nil {
			return err
		}

		for _, row := range rows {
			orderID := orderIDsByNum[row.ID]
			accepted = append(accepted, orderID)
			if err := tx.Exec(`
				UPDATE orders
				SET worker_id = ?, status = 'picked_up', accepted_at = CURRENT_TIMESTAMP, picked_up_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, workerIDUint, row.ID).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				INSERT INTO batch_orders (order_id, batch_id, user_id, status, pickup_time, delivery_time, delivery_address, contact_name, contact_phone, weight, created_at, updated_at)
				VALUES (?, ?, ?, 'Picked Up', CURRENT_TIMESTAMP, NULL, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT (batch_id, order_id) DO UPDATE SET
					user_id = EXCLUDED.user_id,
					status = EXCLUDED.status,
					pickup_time = EXCLUDED.pickup_time,
					delivery_address = EXCLUDED.delivery_address,
					contact_name = EXCLUDED.contact_name,
					contact_phone = EXCLUDED.contact_phone,
					weight = EXCLUDED.weight,
					updated_at = CURRENT_TIMESTAMP
			`, fmt.Sprintf("ord-%d", row.ID), batchID, workerIDUint, firstNonEmpty(row.Address, row.DropArea), firstNonEmpty(row.CustomerName, "Customer"), firstNonEmpty(row.CustomerPhone, "N/A"), normalizedWeight(row.PackageWeightKg)).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "batch_not_found_or_not_assignable") {
			c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
			return
		}
		c.JSON(500, gin.H{"error": "batch_pickup_failed"})
		return
	}

	refreshBatchCache(availableBatchCacheScope)
	refreshBatchCache(workerID)

	c.JSON(200, gin.H{
		"message":          "batch_picked_up",
		"batchId":          batchID,
		"acceptedOrderIds": accepted,
	})
}

type deliverBatchRequest struct {
	DeliveryCode string `json:"deliveryCode"`
}

func DeliverBatch(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	var req deliverBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid_batch_delivery_request"})
		return
	}

	batchID := strings.TrimSpace(c.Param("batch_id"))
	if batchID == "" {
		c.JSON(400, gin.H{"error": "invalid_batch_id"})
		return
	}

	if !HasDB() {
		store.mu.Lock()
		now := nowISO()
		providedCode := strings.TrimSpace(req.DeliveryCode)
		updated := 0
		earningsDelta := 0.0
		deliveredOrderID := ""
		remainingOrders := 0
		zoneAPartialProgress := false

		if providedCode == "" {
			store.mu.Unlock()
			c.JSON(400, gin.H{"error": "incorrect_delivery_code"})
			return
		}

		for idx, order := range store.data.Orders {
			if fmt.Sprintf("%v", order["worker_id"]) != workerID {
				continue
			}
			status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", order["status"])))
			if status != "picked_up" && status != "accepted" {
				continue
			}

			orderID := fmt.Sprintf("%v", order["order_id"])
			if providedCode == deliveryCodeFromOrderID(orderID) {
				order["status"] = "delivered"
				order["delivered_at"] = now
				order["updated_at"] = now
				store.data.Orders[idx] = order
				updated++
				deliveredOrderID = orderID
				zoneAPartialProgress = true
				switch earning := order["earning_inr"].(type) {
				case float64:
					earningsDelta += earning
				case float32:
					earningsDelta += float64(earning)
				case int:
					earningsDelta += float64(earning)
				case int64:
					earningsDelta += float64(earning)
				case uint:
					earningsDelta += float64(earning)
				}
				continue
			}
		}

		if updated == 0 {
			if providedCode != deliveryCodeFromBatchID(batchID) {
				store.mu.Unlock()
				c.JSON(400, gin.H{"error": "incorrect_delivery_code"})
				return
			}

			for idx, order := range store.data.Orders {
				if fmt.Sprintf("%v", order["worker_id"]) != workerID {
					continue
				}
				status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", order["status"])))
				if status != "picked_up" && status != "accepted" {
					continue
				}
				order["status"] = "delivered"
				order["delivered_at"] = now
				order["updated_at"] = now
				store.data.Orders[idx] = order
				updated++
				switch earning := order["earning_inr"].(type) {
				case float64:
					earningsDelta += earning
				case float32:
					earningsDelta += float64(earning)
				case int:
					earningsDelta += float64(earning)
				case int64:
					earningsDelta += float64(earning)
				case uint:
					earningsDelta += float64(earning)
				}
			}
		}

		for _, order := range store.data.Orders {
			if fmt.Sprintf("%v", order["worker_id"]) != workerID {
				continue
			}
			status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", order["status"])))
			if status == "picked_up" || status == "accepted" {
				remainingOrders++
			}
		}

		store.mu.Unlock()
		if updated == 0 {
			c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
			return
		}
		store.mu.Lock()
		if earningsDelta > 0 {
			if current, ok := store.data.Earnings["this_week_actual"].(int); ok {
				store.data.Earnings["this_week_actual"] = current + int(earningsDelta)
			}
			if profile, ok := store.data.WorkerProfiles[workerID]; ok {
				if current, ok := profile["total_earnings_lifetime"].(float64); ok {
					profile["total_earnings_lifetime"] = current + earningsDelta
				} else {
					profile["total_earnings_lifetime"] = earningsDelta
				}
			}
		}
		store.mu.Unlock()
		refreshBatchCache(workerID)
		if zoneAPartialProgress {
			c.JSON(200, gin.H{
				"message":         "order_delivered_in_batch",
				"batchId":         batchID,
				"orderId":         deliveredOrderID,
				"remainingOrders": remainingOrders,
				"batchCompleted":  false,
			})
			return
		}
		c.JSON(200, gin.H{"message": "batch_delivered", "batchId": batchID, "batchCompleted": true})
		return
	}

	workerIDUint, parseErr := parseWorkerID(workerID)
	if parseErr != nil {
		c.JSON(400, gin.H{"error": "invalid_worker_id"})
		return
	}

	batchCompleted := false
	zoneAPartialProgress := false
	deliveredOrderID := ""
	remainingOrders := 0

	err := workerDB.Transaction(func(tx *gorm.DB) error {
		type batchRow struct {
			BatchID         string  `gorm:"column:batch_id"`
			PickupUserID    *uint   `gorm:"column:pickup_user_id"`
			ZoneLevel       string  `gorm:"column:zone_level"`
			Status          string  `gorm:"column:status"`
			BatchEarningINR float64 `gorm:"column:batch_earning_inr"`
			EarningsPosted  bool    `gorm:"column:earnings_posted"`
		}
		var batch batchRow
		if err := tx.Raw(`SELECT batch_id, pickup_user_id, zone_level, status, batch_earning_inr, earnings_posted FROM batches WHERE batch_id = ? LIMIT 1`, batchID).Scan(&batch).Error; err != nil {
			return err
		}
		if batch.BatchID == "" {
			return fmt.Errorf("batch_not_found_or_not_assignable")
		}
		if strings.EqualFold(strings.TrimSpace(batch.Status), "assigned") {
			return fmt.Errorf("batch_not_picked_up")
		}
		if batch.PickupUserID == nil || *batch.PickupUserID != workerIDUint {
			return fmt.Errorf("batch_not_found_or_not_assignable")
		}

		providedCode := strings.TrimSpace(req.DeliveryCode)
		if providedCode == "" {
			return fmt.Errorf("incorrect_delivery_code")
		}
		isZoneA := strings.EqualFold(strings.TrimSpace(batch.ZoneLevel), "A")

		if !isZoneA {
			expectedBatchCode := deliveryCodeFromBatchID(batchID)
			if providedCode != expectedBatchCode {
				return fmt.Errorf("incorrect_delivery_code")
			}
		}

		if strings.EqualFold(strings.TrimSpace(batch.Status), "delivered") {
			return fmt.Errorf("batch_already_delivered")
		}
		if !strings.EqualFold(strings.TrimSpace(batch.Status), "picked up") && !strings.EqualFold(strings.TrimSpace(batch.Status), "picked_up") {
			return fmt.Errorf("batch_not_found_or_not_assignable")
		}

		if isZoneA {
			type batchOrderCodeRow struct {
				OrderID string `gorm:"column:order_id"`
				Status  string `gorm:"column:status"`
			}
			orderRows := make([]batchOrderCodeRow, 0)
			if err := tx.Raw(`SELECT order_id, COALESCE(status, '') AS status FROM batch_orders WHERE batch_id = ?`, batchID).Scan(&orderRows).Error; err != nil {
				return err
			}

			targetOrderID := ""
			targetAlreadyDelivered := false
			for _, row := range orderRows {
				if providedCode != deliveryCodeFromOrderID(row.OrderID) {
					continue
				}
				targetOrderID = row.OrderID
				statusLower := strings.ToLower(strings.TrimSpace(row.Status))
				targetAlreadyDelivered = statusLower == "delivered"
				break
			}

			if targetOrderID == "" {
				return fmt.Errorf("incorrect_delivery_code")
			}
			if targetAlreadyDelivered {
				return fmt.Errorf("order_already_delivered")
			}

			orderNumID, parseOrderErr := parseOrderID(targetOrderID)
			if parseOrderErr != nil {
				return fmt.Errorf("incorrect_delivery_code")
			}

			orderUpdate := tx.Exec(`
				UPDATE orders
				SET status = 'delivered', delivered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
				  AND worker_id = ?
				  AND LOWER(TRIM(COALESCE(status, ''))) IN ('picked_up', 'accepted')
			`, orderNumID, workerIDUint)
			if orderUpdate.Error != nil {
				return orderUpdate.Error
			}
			if orderUpdate.RowsAffected == 0 {
				return fmt.Errorf("order_already_delivered")
			}

			var deliveredTipINR float64
			if err := tx.Raw(`SELECT COALESCE(tip_inr, 0) FROM orders WHERE id = ? AND worker_id = ?`, orderNumID, workerIDUint).Scan(&deliveredTipINR).Error; err != nil {
				return err
			}
			if err := applyWorkerEarningsIncrement(tx, workerIDUint, float64(totalDeliveryEarningINR(deliveredTipINR))); err != nil {
				return err
			}

			if err := tx.Exec(`
				UPDATE batch_orders
				SET status = 'Delivered', delivery_time = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
				WHERE batch_id = ? AND order_id = ?
			`, batchID, targetOrderID).Error; err != nil {
				return err
			}

			if err := tx.Raw(`
				SELECT COUNT(*)
				FROM batch_orders
				WHERE batch_id = ?
				  AND LOWER(TRIM(COALESCE(status, ''))) <> 'delivered'
			`, batchID).Scan(&remainingOrders).Error; err != nil {
				return err
			}

			deliveredOrderID = targetOrderID
			if remainingOrders > 0 {
				zoneAPartialProgress = true
				return nil
			}
		}

		if err := tx.Exec(`
			UPDATE batches
			SET status = 'Delivered', delivery_time = CURRENT_TIMESTAMP, earnings_posted = TRUE, updated_at = CURRENT_TIMESTAMP
			WHERE batch_id = ?
		`, batchID).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE orders
			SET status = 'delivered', delivered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			FROM batch_orders bo
			WHERE bo.order_id = CONCAT('ord-', orders.id)
			  AND bo.batch_id = ?
			  AND orders.worker_id = ?
			  AND LOWER(TRIM(COALESCE(orders.status, ''))) IN ('picked_up', 'accepted')
		`, batchID, workerIDUint).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE batch_orders
			SET status = 'Delivered', delivery_time = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE batch_id = ?
		`, batchID).Error; err != nil {
			return err
		}

		if batch.EarningsPosted {
			batchCompleted = true
			return nil
		}

		if isZoneA {
			batchCompleted = true
			return nil
		}

		if err := applyWorkerEarningsIncrement(tx, workerIDUint, batch.BatchEarningINR); err != nil {
			return err
		}

		batchCompleted = true
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "incorrect_delivery_code") {
			c.JSON(400, gin.H{"error": "incorrect_delivery_code"})
			return
		}
		if strings.Contains(err.Error(), "order_already_delivered") {
			c.JSON(409, gin.H{"error": "order_already_delivered"})
			return
		}
		if strings.Contains(err.Error(), "batch_already_delivered") {
			c.JSON(409, gin.H{"error": "batch_already_delivered"})
			return
		}
		if strings.Contains(err.Error(), "batch_not_picked_up") {
			c.JSON(409, gin.H{"error": "batch_not_picked_up"})
			return
		}
		if strings.Contains(err.Error(), "batch_not_found_or_not_assignable") {
			c.JSON(404, gin.H{"error": "batch_not_found_or_not_assignable"})
			return
		}
		c.JSON(500, gin.H{"error": "batch_delivery_failed"})
		return
	}

	refreshBatchCache(workerID)
	refreshBatchCache(availableBatchCacheScope)
	if zoneAPartialProgress {
		c.JSON(200, gin.H{
			"message":         "order_delivered_in_batch",
			"batchId":         batchID,
			"orderId":         deliveredOrderID,
			"remainingOrders": remainingOrders,
			"batchCompleted":  false,
		})
		return
	}

	c.JSON(200, gin.H{"message": "batch_delivered", "batchId": batchID, "batchCompleted": batchCompleted})
}

func applyWorkerEarningsIncrement(tx *gorm.DB, workerID uint, amount float64) error {
	if amount <= 0 {
		return nil
	}

	if err := tx.Exec(`
		UPDATE worker_profiles
		SET total_earnings_lifetime = COALESCE(total_earnings_lifetime, 0) + ?, updated_at = CURRENT_TIMESTAMP
		WHERE worker_id = ?
	`, amount, workerID).Error; err != nil {
		return err
	}

	if err := tx.Exec(`
		INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings, claim_eligible)
		VALUES (
			?,
			DATE_TRUNC('week', CURRENT_DATE)::date,
			(DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '6 days')::date,
			?,
			FALSE
		)
		ON CONFLICT (worker_id, week_start) DO UPDATE SET
			total_earnings = weekly_earnings_summary.total_earnings + EXCLUDED.total_earnings,
			week_end = EXCLUDED.week_end,
			updated_at = CURRENT_TIMESTAMP
	`, workerID, amount).Error; err != nil {
		return err
	}

	return nil
}

func snapshotFieldString(snapshot gin.H, field string) string {
	if snapshot == nil {
		return ""
	}
	value, _ := snapshot[field].(string)
	return strings.TrimSpace(value)
}

func sameZoneRoute(fromCity, toCity string) bool {
	return strings.EqualFold(strings.TrimSpace(fromCity), strings.TrimSpace(toCity))
}

func matchesWorkerFromZone(scope workerOrderScope, fromCity string) bool {
	fromLower := strings.ToLower(strings.TrimSpace(fromCity))
	if fromLower == "" {
		return false
	}

	zoneNameLower := strings.ToLower(strings.TrimSpace(scope.ZoneName))
	zoneCityLower := strings.ToLower(strings.TrimSpace(scope.ZoneCity))
	if zoneNameLower == "" && zoneCityLower == "" {
		return true
	}

	return (zoneNameLower != "" && fromLower == zoneNameLower) || (zoneCityLower != "" && fromLower == zoneCityLower)
}

func filterAssignedBatchesForScope(batches []gin.H, scope workerOrderScope) []gin.H {
	if len(batches) == 0 {
		return []gin.H{}
	}

	filtered := make([]gin.H, 0, len(batches))
	for _, batch := range batches {
		fromCity := snapshotFieldString(batch, "fromCity")
		toCity := snapshotFieldString(batch, "toCity")
		if !sameZoneRoute(fromCity, toCity) {
			continue
		}
		if !matchesWorkerFromZone(scope, fromCity) {
			continue
		}
		filtered = append(filtered, batch)
	}

	return filtered
}

func filterAvailableBatchesForScope(batches []gin.H, scope workerOrderScope) []gin.H {
	if len(batches) == 0 {
		return []gin.H{}
	}

	filtered := make([]gin.H, 0, len(batches))
	for _, batch := range batches {
		fromCity := snapshotFieldString(batch, "fromCity")
		toCity := snapshotFieldString(batch, "toCity")
		if sameZoneRoute(fromCity, toCity) {
			continue
		}
		if !matchesWorkerFromZone(scope, fromCity) {
			continue
		}
		filtered = append(filtered, batch)
	}

	return filtered
}

func GetAssignedBatches(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	batches := listCachedSnapshotsByStatus(workerID, batchStatusAllowedForAssigned)
	if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
		scope := getWorkerOrderScope(workerIDUint)
		batches = filterAssignedBatchesForScope(batches, scope)
	} else {
		batches = filterAssignedBatchesForScope(batches, workerOrderScope{})
	}
	c.JSON(200, gin.H{"batches": sanitizeBatchSnapshotsForWorker(batches)})
}

func GetAvailableBatches(c *gin.Context) {
	if HasDB() {
		refreshBatchCache(availableBatchCacheScope)
	}
	limitStr := c.DefaultQuery("limit", "100")
	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	batches := listCachedSnapshotsByStatus(availableBatchCacheScope, batchStatusAllowedForAvailable)
	if workerID, ok := optionalAuthWorkerID(c); ok {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			scope := getWorkerOrderScope(workerIDUint)
			batches = filterAvailableBatchesForScope(batches, scope)
		} else {
			batches = filterAvailableBatchesForScope(batches, workerOrderScope{})
		}
	}
	if len(batches) > limit {
		batches = batches[:limit]
	}
	c.JSON(200, gin.H{"batches": sanitizeBatchSnapshotsForWorker(batches)})
}

func GetDeliveredBatches(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	if HasDB() {
		refreshBatchCache(workerID)
	}
	batches := listCachedSnapshotsByStatus(workerID, batchStatusAllowedForDelivered)
	c.JSON(200, gin.H{"batches": sanitizeBatchSnapshotsForWorker(batches)})
}
