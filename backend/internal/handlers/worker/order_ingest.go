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

func bodyStringFromKeys(body map[string]any, fallback string, keys ...string) string {
	for _, key := range keys {
		if value, ok := body[key]; ok {
			if result := bodyString(map[string]any{key: value}, key, ""); strings.TrimSpace(result) != "" {
				return strings.TrimSpace(result)
			}
		}
	}
	return fallback
}

func bodyFloatFromKeys(body map[string]any, fallback float64, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := body[key]; ok {
			switch n := value.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			case uint:
				return float64(n)
			case string:
				trimmed := strings.TrimSpace(n)
				if trimmed == "" {
					continue
				}
				if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
					return parsed
				}
			}
		}
	}
	return fallback
}

func bodyUintFromKeys(body map[string]any, fallback uint, keys ...string) uint {
	for _, key := range keys {
		if value, ok := body[key]; ok {
			switch n := value.(type) {
			case float64:
				return uint(n)
			case int:
				return uint(n)
			case uint:
				return n
			case string:
				parsed, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64)
				if err == nil {
					return uint(parsed)
				}
			}
		}
	}
	return fallback
}

func bodyStringSliceFromKeys(body map[string]any, fallback []string, keys ...string) []string {
	for _, key := range keys {
		if value, ok := body[key]; ok && value != nil {
			return bodyStringSlice(map[string]any{key: value}, key, fallback)
		}
	}
	return append([]string(nil), fallback...)
}

func normalizeZoneLevelValue(raw, fromCity, toCity, fromState, toState string) string {
	compact := strings.ToUpper(strings.TrimSpace(raw))
	compact = strings.NewReplacer("_", "", "-", "", " ", "").Replace(compact)

	switch compact {
	case "A", "ZONEA", "LOCAL", "SAMECITY":
		return "A"
	case "B", "ZONEB", "INTRAZONE", "INTRASTATE":
		return "B"
	case "C", "ZONEC", "INTERSTATE":
		return "C"
	}

	inferred := strings.ToUpper(strings.TrimSpace(inferBatchZoneLevel(fromCity, toCity, fromState, toState)))
	if inferred == "A" || inferred == "B" || inferred == "C" {
		return inferred
	}
	return "A"
}

func zoneTypeFromZoneLevel(zoneLevel string) string {
	switch strings.ToUpper(strings.TrimSpace(zoneLevel)) {
	case "A":
		return "same-city"
	case "B":
		return "intra-state"
	case "C":
		return "inter-state"
	default:
		return "unknown"
	}
}

func zoneLevelFromZoneType(zoneType string, fromCity, toCity, fromState, toState string) string {
	compact := strings.ToUpper(strings.TrimSpace(zoneType))
	compact = strings.NewReplacer("_", "", "-", "", " ", "").Replace(compact)

	switch compact {
	case "A", "ZONEA", "LOCAL", "SAMECITY":
		return "A"
	case "B", "ZONEB", "INTRAZONE", "INTRASTATE":
		return "B"
	case "C", "ZONEC", "INTERSTATE":
		return "C"
	}

	return normalizeZoneLevelValue("", fromCity, toCity, fromState, toState)
}

func deriveZoneRoutePath(zoneLevel string) []string {
	switch strings.ToUpper(strings.TrimSpace(zoneLevel)) {
	case "B":
		return []string{"B", "A"}
	case "C":
		return []string{"C", "B", "A"}
	default:
		return []string{"A"}
	}
}

func generateUniqueOrderID(prefix string) string {
	seen := map[string]struct{}{}

	store.mu.RLock()
	for _, order := range store.data.Orders {
		id := strings.TrimSpace(fmt.Sprintf("%v", order["order_id"]))
		if id != "" {
			seen[id] = struct{}{}
		}
	}
	store.mu.RUnlock()

	for index := len(seen) + 1; ; index++ {
		candidate := nextID(prefix, index)
		if _, ok := seen[candidate]; !ok {
			return candidate
		}
	}
}

func parseDemoOrderPayload(body map[string]any) (order map[string]any, workerID uint) {
	orderID := bodyStringFromKeys(body, "", "order_id")
	if orderID == "" {
		orderID = generateUniqueOrderID("ord")
	}

	customerName := bodyStringFromKeys(body, "Unknown Customer", "customer_name")
	customerID := bodyStringFromKeys(body, "cust-unknown", "customer_id")
	customerContact := bodyStringFromKeys(body, "", "customer_contact_number")
	address := bodyStringFromKeys(body, "Unknown Address", "address")
	paymentMethod := strings.ToLower(bodyStringFromKeys(body, "cod", "payment_method"))
	orderValue := bodyFloatFromKeys(body, bodyFloatFromKeys(body, 0, "payment_amount"), "order_value")
	paymentAmount := bodyFloatFromKeys(body, orderValue, "payment_amount")
	packageSize := strings.ToLower(bodyStringFromKeys(body, "medium", "package_size"))
	packageWeightKg := bodyFloatFromKeys(body, 1.0, "package_weight_kg")
	zoneID := bodyUintFromKeys(body, 1, "zone_id")
	fromCity := bodyStringFromKeys(body, bodyStringFromKeys(body, "Tambaram", "pickup_area"), "from_city")
	toCity := bodyStringFromKeys(body, bodyStringFromKeys(body, "Velachery", "drop_area"), "to_city")
	fromState := bodyStringFromKeys(body, "", "from_state")
	toState := bodyStringFromKeys(body, "", "to_state")
	fromLat := bodyFloatFromKeys(body, 0, "from_lat")
	fromLon := bodyFloatFromKeys(body, 0, "from_lon")
	toLat := bodyFloatFromKeys(body, 0, "to_lat")
	toLon := bodyFloatFromKeys(body, 0, "to_lon")
	pickupArea := bodyStringFromKeys(body, fromCity, "pickup_area")
	dropArea := bodyStringFromKeys(body, toCity, "drop_area")
	distanceKm := bodyFloatFromKeys(body, 3.1, "distance_km")
	tipInr := bodyFloatFromKeys(body, 0, "tip_inr")
	zoneLevel := normalizeZoneLevelValue(
		bodyStringFromKeys(body, "", "zone_level"),
		fromCity,
		toCity,
		fromState,
		toState,
	)
	if zoneType := bodyStringFromKeys(body, "", "zone_type"); zoneType != "" {
		zoneLevel = zoneLevelFromZoneType(zoneType, fromCity, toCity, fromState, toState)
	}
	zoneType := zoneTypeFromZoneLevel(zoneLevel)
	zoneRoutePath := bodyStringSliceFromKeys(body, nil, "zone_route_path", "zone_path")
	if len(zoneRoutePath) == 0 {
		zoneRoutePath = deriveZoneRoutePath(zoneLevel)
	}
	deliveryFeeInr := bodyFloatFromKeys(body, 0, "delivery_fee_inr")
	if deliveryFeeInr <= 0 {
		deliveryFeeInr = float64(computeZoneRouteDeliveryFee(zoneRoutePath))
	}
	status := strings.ToLower(bodyStringFromKeys(body, "assigned", "status"))
	workerID = bodyUintFromKeys(body, 0, "worker_id")
	vehicleType := bodyStringFromKeys(body, "", "vehicle_type")
	vehicleCapacity := bodyInt(body, "vehicle_capacity", 0)
	allowedZones := bodyStringFromKeys(body, "", "allowed_zones")
	assignedAt := bodyStringFromKeys(body, nowISO(), "assigned_at")
	source := bodyStringFromKeys(body, "fake-publisher", "source")

	order = map[string]any{
		"order_id":                orderID,
		"customer_name":           customerName,
		"customer_id":             customerID,
		"customer_contact_number": customerContact,
		"address":                 address,
		"payment_method":          paymentMethod,
		"order_value":             orderValue,
		"payment_amount":          paymentAmount,
		"package_size":            packageSize,
		"package_weight_kg":       packageWeightKg,
		"status":                  status,
		"zone_id":                 zoneID,
		"zone_level":              zoneLevel,
		"zone_type":               zoneType,
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
		"assigned_at":             assignedAt,
		"source":                  source,
	}

	if workerID != 0 {
		order["worker_id"] = fmt.Sprintf("%d", workerID)
	}

	return order, workerID
}

// IngestDemoOrder ingests a fake order payload and stores it in app state.
func IngestDemoOrder(c *gin.Context) {
	body := parseBody(c)
	order, workerID := parseDemoOrderPayload(body)
	orderID := fmt.Sprintf("%v", order["order_id"])
	customerName := fmt.Sprintf("%v", order["customer_name"])
	customerContact := fmt.Sprintf("%v", order["customer_contact_number"])
	orderValue := bodyFloat(order, "order_value", 0)
	packageSize := fmt.Sprintf("%v", order["package_size"])
	packageWeightKg := bodyFloat(order, "package_weight_kg", 0)
	zoneID := bodyUint(order, "zone_id", 1)
	fromCity := fmt.Sprintf("%v", order["from_city"])
	toCity := fmt.Sprintf("%v", order["to_city"])
	fromState := fmt.Sprintf("%v", order["from_state"])
	toState := fmt.Sprintf("%v", order["to_state"])
	fromLat := bodyFloat(order, "from_lat", 0)
	fromLon := bodyFloat(order, "from_lon", 0)
	toLat := bodyFloat(order, "to_lat", 0)
	toLon := bodyFloat(order, "to_lon", 0)
	pickupArea := fmt.Sprintf("%v", order["pickup_area"])
	dropArea := fmt.Sprintf("%v", order["drop_area"])
	distanceKm := bodyFloat(order, "distance_km", 0)
	tipInr := bodyFloat(order, "tip_inr", 0)
	zoneRoutePath := bodyStringSlice(order, "zone_route_path", []string{"A"})
	vehicleType := fmt.Sprintf("%v", order["vehicle_type"])
	vehicleCapacity := bodyInt(order, "vehicle_capacity", 0)
	allowedZones := fmt.Sprintf("%v", order["allowed_zones"])
	deliveryFeeInr := bodyFloat(order, "delivery_fee_inr", 0)
	status := fmt.Sprintf("%v", order["status"])

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

	updatedExisting := false
	store.mu.Lock()

	for idx, existing := range store.data.Orders {
		if existing["order_id"] == orderID {
			store.data.Orders[idx] = order
			updatedExisting = true
			break
		}
	}

	if !updatedExisting {
		store.data.Orders = append([]map[string]any{order}, store.data.Orders...)
		store.data.Notifications = append([]map[string]any{{
			"id":         nextID("ntf", len(store.data.Notifications)),
			"type":       "order_ingested",
			"title":      "New order received",
			"body":       fmt.Sprintf("%s for %s ingested", orderID, customerName),
			"created_at": nowISO(),
			"read":       false,
		}}, store.data.Notifications...)
	}
	store.mu.Unlock()

	refreshBatchSnapshotsForOrder(order)
	scheduleBatchMaterialization(availableBatchCacheScope, order)
	if workerID != 0 {
		scheduleBatchMaterialization(fmt.Sprintf("%d", workerID), order)
	}

	if updatedExisting {
		c.JSON(200, gin.H{"message": "order_updated", "order": order})
		return
	}

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
