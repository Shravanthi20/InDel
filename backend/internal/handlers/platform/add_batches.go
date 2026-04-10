package platform

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type addBatchesRequest struct {
	Count     int     `json:"count"`
	ZoneID    uint    `json:"zone_id"`
	ZoneLevel string  `json:"zone_level"`
	FromCity  string  `json:"from_city"`
	ToCity    string  `json:"to_city"`
	FromState string  `json:"from_state"`
	ToState   string  `json:"to_state"`
	Distance  float64 `json:"distance_km"`
}

type zoneCityRow struct {
	ID    uint   `gorm:"column:id"`
	City  string `gorm:"column:city"`
	State string `gorm:"column:state"`
	Name  string `gorm:"column:name"`
}

func normalizeZoneLevel(level string) string {
	upper := strings.ToUpper(strings.TrimSpace(level))
	switch upper {
	case "A", "B", "C":
		return upper
	default:
		return "A"
	}
}

func resolveRoute(req addBatchesRequest, zones []zoneCityRow) (zoneLevel, fromCity, toCity, fromState, toState string) {
	zoneLevel = normalizeZoneLevel(req.ZoneLevel)
	fromCity = strings.TrimSpace(req.FromCity)
	toCity = strings.TrimSpace(req.ToCity)
	fromState = strings.TrimSpace(req.FromState)
	toState = strings.TrimSpace(req.ToState)

	if fromCity != "" && toCity != "" {
		if zoneLevel == "A" {
			toCity = fromCity
		}
		if zoneLevel == "B" && fromState == "" {
			fromState = toState
		}
		if zoneLevel == "B" && toState == "" {
			toState = fromState
		}
		return zoneLevel, fromCity, toCity, fromState, toState
	}

	if len(zones) == 0 {
		return zoneLevel, "Tambaram", "Tambaram", "", ""
	}

	for _, z := range zones {
		if strings.TrimSpace(z.City) == "" {
			continue
		}
		if zoneLevel == "A" {
			return zoneLevel, z.City, z.City, z.State, z.State
		}
	}

	if zoneLevel == "B" {
		for i := 0; i < len(zones); i++ {
			for j := i + 1; j < len(zones); j++ {
				if strings.EqualFold(strings.TrimSpace(zones[i].State), strings.TrimSpace(zones[j].State)) && !strings.EqualFold(strings.TrimSpace(zones[i].City), strings.TrimSpace(zones[j].City)) {
					return zoneLevel, zones[i].City, zones[j].City, zones[i].State, zones[j].State
				}
			}
		}
	}

	if zoneLevel == "C" {
		for i := 0; i < len(zones); i++ {
			for j := i + 1; j < len(zones); j++ {
				if !strings.EqualFold(strings.TrimSpace(zones[i].State), strings.TrimSpace(zones[j].State)) {
					return zoneLevel, zones[i].City, zones[j].City, zones[i].State, zones[j].State
				}
			}
		}
	}

	if len(zones) == 1 {
		return zoneLevel, zones[0].City, zones[0].City, zones[0].State, zones[0].State
	}

	return zoneLevel, zones[0].City, zones[1].City, zones[0].State, zones[1].State
}

func routePathByLevel(level string) string {
	switch normalizeZoneLevel(level) {
	case "C":
		return `[
			"C",
			"B",
			"A"
		]`
	case "B":
		return `[
			"B",
			"A"
		]`
	default:
		return `["A"]`
	}
}

func defaultDistanceByLevel(level string) float64 {
	switch normalizeZoneLevel(level) {
	case "C":
		return 140
	case "B":
		return 42
	default:
		return 8
	}
}

func deliveryFee(distance float64, interState bool) float64 {
	if interState {
		return distance * 2.0
	}
	return distance * 1.2
}

// AddBatches creates demo orders using Go path logic and existing batch packing behavior.
func AddBatches(c *gin.Context) {
	if !hasDB() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db_not_available"})
		return
	}

	var req addBatchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	count := req.Count
	if count <= 0 {
		count = 6
	}
	if count > 60 {
		count = 60
	}

	var workerID uint
	_ = platformDB.Raw("SELECT id FROM users ORDER BY id ASC LIMIT 1").Scan(&workerID).Error
	if workerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no_workers_found"})
		return
	}

	zones := make([]zoneCityRow, 0)
	_ = platformDB.Raw("SELECT id, city, state, name FROM zones ORDER BY id ASC").Scan(&zones).Error

	zoneLevel, fromCity, toCity, fromState, toState := resolveRoute(req, zones)
	if strings.TrimSpace(fromCity) == "" || strings.TrimSpace(toCity) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route_not_resolved"})
		return
	}

	zoneID := req.ZoneID
	if zoneID == 0 {
		if len(zones) > 0 {
			zoneID = zones[0].ID
		} else {
			zoneID = 1
		}
	}

	distanceKm := req.Distance
	if distanceKm <= 0 {
		distanceKm = defaultDistanceByLevel(zoneLevel)
	}

	interState := strings.TrimSpace(fromState) != "" && strings.TrimSpace(toState) != "" && !strings.EqualFold(strings.TrimSpace(fromState), strings.TrimSpace(toState))
	fee := deliveryFee(distanceKm, interState)
	zoneRoutePath := routePathByLevel(zoneLevel)

	weightPattern := []float64{4.8, 3.9, 2.7, 1.6, 0.9, 4.4, 3.1, 2.2, 1.1, 0.7}
	now := time.Now()
	created := 0
	totalWeight := 0.0

	for i := 0; i < count; i++ {
		weight := weightPattern[i%len(weightPattern)]
		orderValue := 200.0 + float64(i*12)
		tip := 10.0
		err := platformDB.Exec(
			`INSERT INTO orders (
				worker_id, zone_id, order_value, status,
				pickup_area, drop_area, distance_km,
				from_city, to_city, from_state, to_state,
				package_size, package_weight_kg,
				tip_inr, delivery_fee_inr, zone_route_path,
				updated_at
			) VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			workerID, zoneID, orderValue,
			fromCity, toCity, distanceKm,
			fromCity, toCity, fromState, toState,
			"medium", weight,
			tip, fee, zoneRoutePath,
			now,
		).Error
		if err != nil {
			continue
		}
		created++
		totalWeight += weight
	}

	if created == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order_create_failed"})
		return
	}

	estimatedBatches := int(math.Ceil(totalWeight / 12.0))
	if estimatedBatches < 1 {
		estimatedBatches = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "batches_added",
		"created_orders":      created,
		"estimated_batches":   estimatedBatches,
		"batch_weight_policy": "target 10-12kg; hard cap 12kg; final batch may be lower",
		"route": gin.H{
			"zone_level": zoneLevel,
			"from_city":  fromCity,
			"to_city":    toCity,
			"from_state": fromState,
			"to_state":   toState,
		},
		"meta": gin.H{
			"zone_id":      zoneID,
			"worker_id":    workerID,
			"distance_km":  distanceKm,
			"delivery_fee": fmt.Sprintf("%.2f", fee),
		},
	})
}
