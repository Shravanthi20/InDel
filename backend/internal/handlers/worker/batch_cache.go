package worker

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var batchMaterializationDelay = loadBatchMaterializationDelay()

const availableBatchCacheScope = "__available__"

func loadBatchMaterializationDelay() time.Duration {
	raw := strings.TrimSpace(os.Getenv("INDEL_BATCH_MATERIALIZATION_DELAY_SECONDS"))
	if raw == "" {
		return 10 * time.Second
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 0 {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func batchGroupKeyFromFields(fromCity, toCity, fromState, toState string) string {
	zoneLevel := inferBatchZoneLevel(fromCity, toCity, fromState, toState)
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s",
		strings.ToLower(strings.TrimSpace(fromCity)),
		strings.ToLower(strings.TrimSpace(toCity)),
		strings.ToLower(strings.TrimSpace(fromState)),
		strings.ToLower(strings.TrimSpace(toState)),
		zoneLevel,
	)
}

func batchGroupKeyFromRow(row batchOrderRow) string {
	return batchGroupKeyFromFields(row.FromCity, row.ToCity, row.FromState, row.ToState)
}

func batchGroupKeyFromOrder(order map[string]any) string {
	fromCity, _ := order["from_city"].(string)
	toCity, _ := order["to_city"].(string)
	fromState, _ := order["from_state"].(string)
	toState, _ := order["to_state"].(string)
	return batchGroupKeyFromFields(fromCity, toCity, fromState, toState)
}

func batchSnapshotKey(snapshot gin.H) string {
	if snapshot == nil {
		return ""
	}
	if key, ok := snapshot["batchKey"].(string); ok {
		return key
	}
	if key, ok := snapshot["batchId"].(string); ok {
		return key
	}
	return ""
}

func batchSnapshotGroupKey(snapshot gin.H) string {
	if snapshot == nil {
		return ""
	}
	if key, ok := snapshot["batchGroupKey"].(string); ok {
		return key
	}
	return batchSnapshotKey(snapshot)
}

func batchStatusAllowedForAvailable(status string) bool {
	switch status {
	case "Assigned":
		return true
	default:
		return false
	}
}

func batchStatusAllowedForAssigned(status string) bool {
	switch status {
	case "Picked Up":
		return true
	default:
		return false
	}
}

func batchStatusAllowedForDelivered(status string) bool {
	switch status {
	case "Delivered":
		return true
	default:
		return false
	}
}

func refreshBatchCache(workerID string) {
	if workerID == "" {
		return
	}
	rows := readBatchRowsForWorker(workerID)
	grouped := groupRowsByBatchKey(rows)

	store.batchMu.Lock()
	defer store.batchMu.Unlock()

	if store.batchCache == nil {
		store.batchCache = map[string]map[string]gin.H{}
	}
	updated := map[string]gin.H{}
	for groupKey, groupRows := range grouped {
		snapshots := rowsToBatches(groupRows, batchStatusFromRows(groupRows, "Assigned"), true)
		if len(snapshots) == 0 {
			continue
		}
		for index := range snapshots {
			snapshots[index]["batchGroupKey"] = groupKey
			snapshots[index]["batchKey"] = fmt.Sprintf("%s#%02d", groupKey, index+1)
			updated[batchSnapshotKey(snapshots[index])] = snapshots[index]
		}
	}
	store.batchCache[workerID] = updated
}

func refreshBatchSnapshotsForOrder(order map[string]any) {
	if order == nil {
		return
	}

	refreshBatchSnapshotForOrder(availableBatchCacheScope, order)
	workerID := strings.TrimSpace(fmt.Sprintf("%v", order["worker_id"]))
	if workerID != "" && workerID != "0" {
		refreshBatchSnapshotForOrder(workerID, order)
	}
}


func readBatchRowsForWorker(workerID string) []batchOrderRow {
	rows := make([]batchOrderRow, 0)

	if HasDB() {
		if workerID == availableBatchCacheScope {
			err := workerDB.Raw(`
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
					COALESCE(o.picked_up_at::text, '') AS picked_up_at,
					COALESCE(o.delivered_at::text, '') AS delivered_at,
					o.created_at::text AS created_at
				FROM orders o
				WHERE o.status IN ('assigned', 'accepted', 'picked_up', 'delivered')
				ORDER BY o.created_at DESC
			`).Scan(&rows).Error
			if err == nil {
				return rows
			}
		}

		workerIDUint, err := parseWorkerID(workerID)
		if err == nil {
			err = workerDB.Raw(`
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
					COALESCE(o.picked_up_at::text, '') AS picked_up_at,
					COALESCE(o.delivered_at::text, '') AS delivered_at,
					o.created_at::text AS created_at
				FROM orders o
				WHERE o.worker_id = ?
				  AND o.status IN ('assigned', 'accepted', 'picked_up', 'delivered')
				ORDER BY o.created_at DESC
			`, workerIDUint).Scan(&rows).Error
			if err == nil {
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
				return rows
			}
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for idx, order := range store.data.Orders {
		orderWorkerID := fmt.Sprintf("%v", order["worker_id"])
		if workerID != "" && orderWorkerID != "" && orderWorkerID != workerID {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", order["status"])))
		if status != "assigned" && status != "accepted" && status != "picked_up" && status != "delivered" {
			continue
		}
		zoneID := uint(0)
		if v, ok := order["zone_id"]; ok {
			switch n := v.(type) {
			case uint:
				zoneID = n
			case int:
				zoneID = uint(n)
			case float64:
				zoneID = uint(n)
			}
		}
		packageWeight := 0.0
		if v, ok := order["package_weight_kg"]; ok {
			switch n := v.(type) {
			case float64:
				packageWeight = n
			case int:
				packageWeight = float64(n)
			}
		}
		rows = append(rows, batchOrderRow{
			ID:              uint(idx + 1),
			ZoneID:          zoneID,
			FromCity:        fmt.Sprintf("%v", order["from_city"]),
			ToCity:          fmt.Sprintf("%v", order["to_city"]),
			FromState:       fmt.Sprintf("%v", order["from_state"]),
			ToState:         fmt.Sprintf("%v", order["to_state"]),
			PickupArea:      fmt.Sprintf("%v", order["pickup_area"]),
			DropArea:        fmt.Sprintf("%v", order["drop_area"]),
			PackageWeightKg: packageWeight,
			Status:          fmt.Sprintf("%v", order["status"]),
			CreatedAt:       fmt.Sprintf("%v", order["created_at"]),
			PickedUpAt:      fmt.Sprintf("%v", order["picked_up_at"]),
			DeliveredAt:     fmt.Sprintf("%v", order["delivered_at"]),
		})
	}
	return rows
}

func groupRowsByBatchKey(rows []batchOrderRow) map[string][]batchOrderRow {
	groups := make(map[string][]batchOrderRow)
	for _, row := range rows {
		key := batchGroupKeyFromRow(row)
		groups[key] = append(groups[key], row)
	}
	return groups
}

func storeBatchSnapshot(workerID string, snapshot gin.H) {
	if workerID == "" || snapshot == nil {
		return
	}
	groupKey := batchSnapshotKey(snapshot)
	if groupKey == "" {
		return
	}

	store.batchMu.Lock()
	defer store.batchMu.Unlock()

	if store.batchCache == nil {
		store.batchCache = map[string]map[string]gin.H{}
	}
	if _, ok := store.batchCache[workerID]; !ok {
		store.batchCache[workerID] = map[string]gin.H{}
	}
	store.batchCache[workerID][groupKey] = snapshot
}

func storeBatchSnapshots(workerID string, snapshots []gin.H) {
	for _, snapshot := range snapshots {
		storeBatchSnapshot(workerID, snapshot)
	}
}

func deleteBatchSnapshot(workerID, groupKey string) {
	if workerID == "" || groupKey == "" {
		return
	}

	store.batchMu.Lock()
	defer store.batchMu.Unlock()
	if store.batchCache == nil {
		return
	}
	if workerBuckets, ok := store.batchCache[workerID]; ok {
		for key, snapshot := range workerBuckets {
			if batchSnapshotGroupKey(snapshot) == groupKey || strings.HasPrefix(key, groupKey+"#") {
				delete(workerBuckets, key)
			}
		}
		if len(workerBuckets) == 0 {
			delete(store.batchCache, workerID)
		}
	}
}

func listCachedSnapshots(workerID string) []gin.H {
	store.batchMu.Lock()
	defer store.batchMu.Unlock()

	workerBuckets, ok := store.batchCache[workerID]
	if !ok || len(workerBuckets) == 0 {
		return nil
	}

	keys := make([]string, 0, len(workerBuckets))
	for key := range workerBuckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		result = append(result, workerBuckets[key])
	}
	return result
}

func listCachedSnapshotsByStatus(workerID string, allow func(string) bool) []gin.H {
	snapshots := listCachedSnapshots(workerID)
	if len(snapshots) == 0 {
		return []gin.H{}
	}
	filtered := make([]gin.H, 0, len(snapshots))
	for _, snapshot := range snapshots {
		status, _ := snapshot["status"].(string)
		if allow(status) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

func refreshBatchSnapshotForWorker(workerID, groupKey string) {
	if workerID == "" || groupKey == "" {
		return
	}

	rows := readBatchRowsForWorker(workerID)
	grouped := groupRowsByBatchKey(rows)
	groupRows, ok := grouped[groupKey]
	if !ok || len(groupRows) == 0 {
		deleteBatchSnapshot(workerID, groupKey)
		return
	}

	snapshots := rowsToBatches(groupRows, "Pending", true)
	if len(snapshots) == 0 {
		deleteBatchSnapshot(workerID, groupKey)
		return
	}
	for index := range snapshots {
		snapshots[index]["batchGroupKey"] = groupKey
		snapshots[index]["batchKey"] = fmt.Sprintf("%s#%02d", groupKey, index+1)
	}
	deleteBatchSnapshot(workerID, groupKey)
	storeBatchSnapshots(workerID, snapshots)
}

func refreshBatchSnapshotForOrder(workerID string, order map[string]any) {
	groupKey := batchGroupKeyFromOrder(order)
	refreshBatchSnapshotForWorker(workerID, groupKey)
}

func findStoredOrder(orderID string) (map[string]any, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, order := range store.data.Orders {
		if fmt.Sprintf("%v", order["order_id"]) == orderID {
			copyOrder := make(map[string]any, len(order))
			for key, value := range order {
				copyOrder[key] = value
			}
			return copyOrder, true
		}
	}
	return nil, false
}

func refreshBatchSnapshotForOrderID(workerID, orderID string) {
	order, ok := findStoredOrder(orderID)
	if !ok {
		return
	}
	refreshBatchSnapshotForOrder(workerID, order)
}

func mirrorStoredOrderStatus(orderID, workerID, newStatus string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for idx, order := range store.data.Orders {
		if fmt.Sprintf("%v", order["order_id"]) != orderID {
			continue
		}
		if workerID != "" {
			order["worker_id"] = workerID
		}
		order["status"] = newStatus
		order["updated_at"] = nowISO()
		if newStatus == "picked_up" {
			order["accepted_at"] = nowISO()
			order["picked_up_at"] = nowISO()
		}
		if newStatus == "delivered" {
			order["delivered_at"] = nowISO()
		}
		store.data.Orders[idx] = order
		return
	}
}

func scheduleBatchMaterialization(workerID string, order map[string]any) {
	if workerID == "" || order == nil {
		return
	}
	groupKey := batchGroupKeyFromOrder(order)
	if groupKey == "" {
		return
	}

	timerKey := workerID + "|" + groupKey
	store.batchMu.Lock()
	if store.batchTimers == nil {
		store.batchTimers = map[string]*time.Timer{}
	}
	if existing, ok := store.batchTimers[timerKey]; ok && existing != nil {
		existing.Stop()
	}
	store.batchTimers[timerKey] = time.AfterFunc(batchMaterializationDelay, func() {
		refreshBatchSnapshotForWorker(workerID, groupKey)
		store.batchMu.Lock()
		delete(store.batchTimers, timerKey)
		store.batchMu.Unlock()
	})
	store.batchMu.Unlock()
}

func clearBatchMaterializationTimers() {
	store.batchMu.Lock()
	defer store.batchMu.Unlock()
	for key, timer := range store.batchTimers {
		if timer != nil {
			timer.Stop()
		}
		delete(store.batchTimers, key)
	}
}
