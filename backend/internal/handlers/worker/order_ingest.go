package worker

import (
	"fmt"
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

// IngestDemoOrder ingests a fake order payload and stores it in app state.
func IngestDemoOrder(c *gin.Context) {
	body := parseBody(c)

	orderID := bodyString(body, "order_id", nextID("ord", len(store.data.Orders)))
	customerName := bodyString(body, "customer_name", "Unknown Customer")
	customerID := bodyString(body, "customer_id", "cust-unknown")
	customerContact := bodyString(body, "customer_contact_number", "")
	address := bodyString(body, "address", "Unknown Address")
	paymentMethod := bodyString(body, "payment_method", "cod")
	paymentAmount := bodyFloat(body, "payment_amount", 0)
	packageSize := bodyString(body, "package_size", "medium")
	packageWeightKg := bodyFloat(body, "package_weight_kg", 1.0)

	if customerContact == "" {
		c.JSON(400, gin.H{"error": "customer_contact_number_required"})
		return
	}

	order := map[string]any{
		"order_id":                orderID,
		"customer_name":           customerName,
		"customer_id":             customerID,
		"customer_contact_number": customerContact,
		"address":                 address,
		"payment_method":          strings.ToLower(paymentMethod),
		"payment_amount":          paymentAmount,
		"package_size":            strings.ToLower(packageSize),
		"package_weight_kg":       packageWeightKg,
		"status":                  bodyString(body, "status", "assigned"),
		"assigned_at":             bodyString(body, "assigned_at", nowISO()),
		"source":                  bodyString(body, "source", "fake-publisher"),
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
