package platform

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIdempotency(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	zoneID := uint(1)
	orderID := "test_order_xyz"

	// Track order first time - should return true
	tracked1 := CheckAndTrackOrder(orderID, zoneID, false)
	if !tracked1 {
		t.Errorf("Expected first tracking of order %s to succeed, got false", orderID)
	}

	// Track order second time - should be blocked by idempotency lock
	tracked2 := CheckAndTrackOrder(orderID, zoneID, false)
	if tracked2 {
		t.Errorf("Expected duplicate tracking of order %s to be blocked, got true", orderID)
	}
}

func TestDisruptionConfidenceScoring(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	zoneID := uint(2)
	state := getOrCreateZoneState(zoneID)
	
	state.BaselineOrders = 100.0
	state.TotalOrdersEver = 11
	
	for i := 0; i < 50; i++ {
		CheckAndTrackOrder("dummy"+string(rune(i)), zoneID, true)
	}

	state.ActiveSignals["weather_test"] = true

	evaluateDisruption(zoneID, state)
}

func TestMultiSignalConfirmationRules(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()
	
	zoneID := uint(3)
	state := getOrCreateZoneState(zoneID)
	state.BaselineOrders = 100.0 
	state.TotalOrdersEver = 11

	state.mu.Lock()
	state.RecentOrders = state.RecentOrders[:0] // 100% drop logically initially
	state.mu.Unlock()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	GetZoneHealth(c)

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &response)

	found := false
	for _, z := range response.Data {
		if uint(z["zone_id"].(float64)) == zoneID {
			found = true
			if z["status"] != "anomalous_demand" {
				t.Errorf("Expected anomalous_demand for drop without external signal, got %v", z["status"])
			}
		}
	}
	if !found {
		t.Errorf("Zone %d not found in health output", zoneID)
	}

	state.mu.Lock()
	state.ActiveSignals["strike"] = true
	state.mu.Unlock()

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	GetZoneHealth(c2)

	var response2 struct {
		Data []map[string]interface{} `json:"data"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &response2)

	for _, z := range response2.Data {
		if uint(z["zone_id"].(float64)) == zoneID {
			if z["status"] != "disrupted" {
				t.Errorf("Expected disrupted status when drop and signal present, got %v", z["status"])
			}
		}
	}
}

func TestZoneHealthAggregation(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	zoneID := uint(4)
	state := getOrCreateZoneState(zoneID)
	state.mu.Lock()
	state.BaselineOrders = 200.0
	state.mu.Unlock()
	
	for i := 0; i < 200; i++ {
		CheckAndTrackOrder("healthy"+string(rune(i)), zoneID, true)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	GetZoneHealth(c)

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	for _, z := range response.Data {
		if uint(z["zone_id"].(float64)) == zoneID {
			if z["status"] != "healthy" {
				t.Errorf("Expected status 'healthy', got %v", z["status"])
			}
			if z["baseline_orders"] != 200.0 {
				t.Errorf("Expected 200 baseline orders, got %v", z["baseline_orders"])
			}
		}
	}
}
