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

func GetSimulationBatches(c *gin.Context) {
	if _, ok := requireDemoOperationRole(c); !ok {
		return
	}

	if HasDB() {
		refreshBatchCache(availableBatchCacheScope)
	}

	all := append([]gin.H{}, listCachedSnapshots(availableBatchCacheScope)...)
	statusFilter := strings.ToLower(strings.TrimSpace(c.Query("status")))
	if statusFilter != "" {
		filtered := make([]gin.H, 0, len(all))
		for _, batch := range all {
			status, _ := batch["status"].(string)
			if strings.ToLower(strings.ReplaceAll(strings.TrimSpace(status), " ", "_")) == statusFilter {
				filtered = append(filtered, batch)
			}
		}
		all = filtered
	}

	c.JSON(200, gin.H{"batches": all})
}
