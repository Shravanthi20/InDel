package worker

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func sanitizeBatchSnapshotForWorker(snapshot gin.H) gin.H {
	if snapshot == nil {
		return gin.H{}
	}

	clone := gin.H{}
	for key, value := range snapshot {
		switch key {
		case "pickupCode", "deliveryCode":
			continue
		case "orders":
			rawOrders, _ := value.([]gin.H)
			if rawOrders == nil {
				if genericOrders, ok := value.([]interface{}); ok {
					sanitizedOrders := make([]gin.H, 0, len(genericOrders))
					for _, genericOrder := range genericOrders {
						if orderMap, ok := genericOrder.(map[string]interface{}); ok {
							orderClone := gin.H{}
							for orderKey, orderValue := range orderMap {
								if orderKey == "deliveryCode" {
									continue
								}
								orderClone[orderKey] = orderValue
							}
							sanitizedOrders = append(sanitizedOrders, orderClone)
						}
					}
					clone[key] = sanitizedOrders
					continue
				}
			}

			sanitizedOrders := make([]gin.H, 0, len(rawOrders))
			for _, order := range rawOrders {
				orderClone := gin.H{}
				for orderKey, orderValue := range order {
					if orderKey == "deliveryCode" {
						continue
					}
					orderClone[orderKey] = orderValue
				}
				sanitizedOrders = append(sanitizedOrders, orderClone)
			}
			clone[key] = sanitizedOrders
		default:
			clone[key] = value
		}
	}

	return clone
}

func sanitizeBatchSnapshotsForWorker(snapshots []gin.H) []gin.H {
	sanitized := make([]gin.H, 0, len(snapshots))
	for _, snapshot := range snapshots {
		sanitized = append(sanitized, sanitizeBatchSnapshotForWorker(snapshot))
	}
	return sanitized
}

func normalizeSimulationBatchStatus(status string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(status), " ", "_"))
}

func filterSimulationBatchesByStatus(batches []gin.H, statusFilter string) []gin.H {
	if statusFilter == "" {
		return batches
	}

	filtered := make([]gin.H, 0, len(batches))
	for _, batch := range batches {
		status, _ := batch["status"].(string)
		if normalizeSimulationBatchStatus(status) == statusFilter {
			filtered = append(filtered, batch)
		}
	}
	return filtered
}

func GetSimulationBatches(c *gin.Context) {
	if _, ok := requireDemoOperationRole(c); !ok {
		return
	}

	if HasDB() {
		refreshBatchCache(availableBatchCacheScope)
	}

	statusFilter := strings.ToLower(strings.TrimSpace(c.Query("status")))
	available := sanitizeBatchSnapshotsForWorker(listCachedSnapshotsByStatus(availableBatchCacheScope, batchStatusAllowedForAvailable))
	assigned := sanitizeBatchSnapshotsForWorker(listCachedSnapshotsByStatus(availableBatchCacheScope, func(status string) bool {
		return batchStatusAllowedForAssigned(status) || batchStatusAllowedForDelivered(status)
	}))
	all := append([]gin.H{}, available...)
	all = append(all, assigned...)

	if statusFilter != "" {
		available = filterSimulationBatchesByStatus(available, statusFilter)
		assigned = filterSimulationBatchesByStatus(assigned, statusFilter)
		all = filterSimulationBatchesByStatus(all, statusFilter)
	}

	c.JSON(200, gin.H{
		"batches":           all,
		"available_batches": available,
		"assigned_batches":  assigned,
	})
}
