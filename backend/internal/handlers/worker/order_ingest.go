package worker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func bodyFloat(body map[string]any, key string, fallback float64) float64 {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return fallback
	}
}

func bodyUint(body map[string]any, key string, fallback uint) uint {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return uint(n)
	case int:
		return uint(n)
	case uint:
		return n
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64)
		if err != nil {
			return fallback
		}
		return uint(parsed)
	default:
		return fallback
	}
}

func bodyStringSlice(body map[string]any, key string, fallback []string) []string {
	v, ok := body[key]
	if !ok || v == nil {
		return normalizeZoneBandPath(fallback)
	}
	switch val := v.(type) {
	case []any:
		items := make([]string, 0, len(val))
		for _, it := range val {
			items = append(items, fmt.Sprintf("%v", it))
		}
		return normalizeZoneBandPath(items)
	case []string:
		return normalizeZoneBandPath(val)
	case string:
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return normalizeZoneBandPath(fallback)
		}
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			return normalizeZoneBandPath(arr)
		}
		if strings.Contains(trimmed, ">") {
			return normalizeZoneBandPath(strings.Split(trimmed, ">"))
		}
		if strings.Contains(trimmed, ",") {
			return normalizeZoneBandPath(strings.Split(trimmed, ","))
		}
		return normalizeZoneBandPath([]string{trimmed})
	default:
		return normalizeZoneBandPath(fallback)
	}
}

// IngestDemoOrder ingests a fake order payload and stores it in app state.
func IngestDemoOrder(c *gin.Context) {
	body := parseBody(c)

	orderID := bodyString(body, "order_id", nextID("ord", len(store.data.Orders)))
	customerName := bodyString(body, "customer_name", "Unknown Customer")
	customerID := bodyString(body, "customer_id", "cust-unknown")
	customerContact := bodyString(body, "customer_contact_number", "")
	address := bodyString(body, "address", "Unknown Address")
	paymentMethod := bodyString(body, "payment_method", "cod")
	orderValue := bodyFloat(body, "order_value", bodyFloat(body, "payment_amount", 0))
	paymentAmount := bodyFloat(body, "payment_amount", orderValue)
	packageSize := bodyString(body, "package_size", "medium")
	packageWeightKg := bodyFloat(body, "package_weight_kg", 1.0)
	zoneID := bodyUint(body, "zone_id", 1)
	fromCity := bodyString(body, "from_city", bodyString(body, "pickup_area", "Tambaram"))
	toCity := bodyString(body, "to_city", bodyString(body, "drop_area", "Velachery"))
	fromState := bodyString(body, "from_state", "")
	toState := bodyString(body, "to_state", "")
	fromLat := bodyFloat(body, "from_lat", 0)
	fromLon := bodyFloat(body, "from_lon", 0)
	toLat := bodyFloat(body, "to_lat", 0)
	toLon := bodyFloat(body, "to_lon", 0)
	pickupArea := bodyString(body, "pickup_area", "Tambaram")
	dropArea := bodyString(body, "drop_area", "Velachery")
	distanceKm := bodyFloat(body, "distance_km", 3.1)
	tipInr := bodyFloat(body, "tip_inr", 0)
	zoneRoutePath := bodyStringSlice(body, "zone_route_path", []string{"A"})
	vehicleType := bodyString(body, "vehicle_type", "")
	vehicleCapacity := bodyInt(body, "vehicle_capacity", 0)
	allowedZones := bodyString(body, "allowed_zones", "")
	deliveryFeeInr := bodyFloat(body, "delivery_fee_inr", float64(computeZoneRouteDeliveryFee(zoneRoutePath)))
	status := bodyString(body, "status", "assigned")
	workerID := bodyUint(body, "worker_id", 0)

	if customerContact == "" {
		c.JSON(400, gin.H{"error": "customer_contact_number_required"})
		return
	}

	if HasDB() {
		if workerID == 0 {
			type firstUserRow struct {
				ID uint `gorm:"column:id"`
			}
			var u firstUserRow
			_ = workerDB.Raw("SELECT id FROM users ORDER BY id ASC LIMIT 1").Scan(&u).Error
			if u.ID != 0 {
				workerID = u.ID
			}
		}
		if workerID != 0 {
			_ = workerDB.Exec(
				`INSERT INTO orders (
					worker_id, zone_id, order_value,
					package_size, package_weight_kg,
					from_city, to_city, from_state, to_state,
					from_lat, from_lon, to_lat, to_lon,
					status, pickup_area, drop_area, distance_km,
					tip_inr, delivery_fee_inr, zone_route_path,
					vehicle_type, vehicle_capacity, allowed_zones,
					updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
				workerID, zoneID, orderValue,
				strings.ToLower(packageSize), packageWeightKg,
				fromCity, toCity, fromState, toState,
				fromLat, fromLon, toLat, toLon,
				status, pickupArea, dropArea, distanceKm,
				tipInr, deliveryFeeInr, encodeZonePath(zoneRoutePath),
				vehicleType, vehicleCapacity, allowedZones,
			).Error
		}
	}

	order := map[string]any{
		"order_id":                orderID,
		"customer_name":           customerName,
		"customer_id":             customerID,
		"customer_contact_number": customerContact,
		"address":                 address,
		"payment_method":          strings.ToLower(paymentMethod),
		"order_value":             orderValue,
		"payment_amount":          paymentAmount,
		"package_size":            strings.ToLower(packageSize),
		"package_weight_kg":       packageWeightKg,
		"status":                  status,
		"zone_id":                 zoneID,
		"from_city":               fromCity,
		"to_city":                 toCity,
		"from_state":              fromState,
		"to_state":                toState,
		"from_lat":                fromLat,
		"from_lon":                fromLon,
		"to_lat":                  toLat,
		"to_lon":                  toLon,
		"pickup_area":             pickupArea,
		"drop_area":               dropArea,
		"distance_km":             distanceKm,
		"tip_inr":                 tipInr,
		"delivery_fee_inr":        deliveryFeeInr,
		"zone_route_path":         zoneRoutePath,
		"zone_route_display":      zonePathDisplay(zoneRoutePath),
		"vehicle_type":            vehicleType,
		"vehicle_capacity":        vehicleCapacity,
		"allowed_zones":           allowedZones,
		"earning_inr":             totalDeliveryEarningINR(tipInr),
		"assigned_at":             bodyString(body, "assigned_at", nowISO()),
		"source":                  bodyString(body, "source", "fake-publisher"),
	}
	if workerID != 0 {
		order["worker_id"] = fmt.Sprintf("%d", workerID)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for idx, existing := range store.data.Orders {
		if existing["order_id"] == orderID {
			store.data.Orders[idx] = order
			c.JSON(200, gin.H{"message": "order_updated", "order": order})
			return
		}
	}

	store.data.Orders = append([]map[string]any{order}, store.data.Orders...)
	scheduleBatchMaterialization(availableBatchCacheScope, order)
	if workerID != 0 {
		scheduleBatchMaterialization(fmt.Sprintf("%d", workerID), order)
	}
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "order_ingested",
		"title":      "New order received",
		"body":       fmt.Sprintf("%s for %s ingested", orderID, customerName),
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)

	c.JSON(201, gin.H{"message": "order_ingested", "order": order})
}

// SearchDemoOrders searches ingested orders by a query string.
func SearchDemoOrders(c *gin.Context) {
	q := strings.TrimSpace(strings.ToLower(c.Query("query")))
	limit := 20

	store.mu.RLock()
	defer store.mu.RUnlock()

	if q == "" {
		orders := append([]map[string]any{}, store.data.Orders...)
		if len(orders) > limit {
			orders = orders[:limit]
		}
		c.JSON(200, gin.H{"count": len(orders), "orders": orders})
		return
	}

	results := make([]map[string]any, 0)
	for _, order := range store.data.Orders {
		blob := strings.ToLower(fmt.Sprintf("%v %v %v %v %v %v %v %v %v",
			order["order_id"],
			order["customer_name"],
			order["customer_id"],
			order["customer_contact_number"],
			order["address"],
			order["payment_method"],
			order["package_size"],
			order["status"],
			order["source"],
		))
		if strings.Contains(blob, q) {
			results = append(results, order)
		}
		if len(results) >= limit {
			break
		}
	}

	c.JSON(200, gin.H{"count": len(results), "orders": results})
}
