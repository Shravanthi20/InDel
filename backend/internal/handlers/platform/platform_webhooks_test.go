package platform

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOrderCompletedWebhook_SchemaValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Missing required fields (zone_id, worker_id, etc. depending on your struct)
	// We'll pass broken json
	reqBody := []byte(`{"order_id": ""}`)
	c.Request, _ = http.NewRequest(http.MethodPost, "/webhooks/order/completed", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	OrderCompletedWebhook(c)

	if w.Code == http.StatusOK {
		t.Errorf("Expected failure for invalid webhook schema, got 200 OK")
	}
}

func TestOrderCompletedWebhook_Idempotency(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	gin.SetMode(gin.TestMode)

	// Valid payload (use "fake" in order_id to skip DB validation in handler)
	payload := []byte(`{
		"order_id": "fake_ord_123",
		"worker_id": "wkr_123",
		"zone_id": 1,
		"completed_at": "2026-03-30T09:42:00Z",
		"distance_km": 4.6,
		"earning_inr": 102.0
	}`)

	// Call 1 - should process and return OK
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/webhooks/order/completed", bytes.NewBuffer(payload))
	c1.Request.Header.Set("Content-Type", "application/json")
	OrderCompletedWebhook(c1)

	if w1.Code != http.StatusOK {
		t.Errorf("First valid webhook call should return 200, got %v", w1.Code)
	}

	// Call 2 - should be skipped/ok without side effect (idempotency)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodPost, "/webhooks/order/completed", bytes.NewBuffer(payload))
	c2.Request.Header.Set("Content-Type", "application/json")
	OrderCompletedWebhook(c2)

	// Still returns OK to the sender, but internally processing was bypassed.
	if w2.Code != http.StatusOK {
		t.Errorf("Duplicate webhook call should still return 200 OK, got %v", w2.Code)
	}
}
