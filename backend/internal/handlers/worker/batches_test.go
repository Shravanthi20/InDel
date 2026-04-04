package worker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAuthedBatchAcceptContext(batchID string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/batches/"+batchID+"/accept", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = gin.Params{{Key: "batch_id", Value: batchID}}
	return c, w
}

func newAuthedBatchDeliverContext(batchID string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/batches/"+batchID+"/deliver", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = gin.Params{{Key: "batch_id", Value: batchID}}
	return c, w
}

func TestAcceptBatchSetsOrdersToPickedUp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	batchID := "BATCH_TEST"
	c, w := newAuthedBatchAcceptContext(batchID, map[string]any{
		"orderIds":   []string{"ord-001", "ord-002"},
		"pickupCode": pickupCodeFromBatchID(batchID),
	})
	AcceptBatch(c)

	if w.Code != 200 {
		t.Fatalf("AcceptBatch status = %d, want 200", w.Code)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, order := range store.data.Orders {
		status, _ := order["status"].(string)
		if status != "picked_up" {
			t.Fatalf("order %v status = %q, want picked_up", order["order_id"], status)
		}
	}
}

func TestAcceptBatchRejectsIncorrectPickupCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	c, w := newAuthedBatchAcceptContext("BATCH_TEST", map[string]any{
		"orderIds":   []string{"ord-001"},
		"pickupCode": "9999",
	})
	AcceptBatch(c)

	if w.Code != 400 {
		t.Fatalf("AcceptBatch status = %d, want 400", w.Code)
	}
}

func TestBatchStatusFromRowsTransitionsToPickedUp(t *testing.T) {
	rows := []batchOrderRow{
		{Status: "accepted"},
		{Status: "picked_up"},
	}

	got := batchStatusFromRows(rows, "pending")
	if got != "Picked Up" {
		t.Fatalf("batchStatusFromRows() = %q, want %q", got, "Picked Up")
	}
}

func TestDeliverBatchMarksPickedUpOrdersDelivered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	store.mu.Lock()
	store.data.Orders = []map[string]any{
		{"order_id": "ord-001", "worker_id": "worker-001", "status": "picked_up", "updated_at": "before"},
		{"order_id": "ord-002", "worker_id": "worker-001", "status": "picked_up", "updated_at": "before"},
		{"order_id": "ord-003", "worker_id": "worker-001", "status": "assigned", "updated_at": "before"},
	}
	store.mu.Unlock()

	batchID := "BATCH_TEST"
	c, w := newAuthedBatchDeliverContext(batchID, map[string]any{
		"deliveryCode": deliveryCodeFromBatchID(batchID),
	})
	DeliverBatch(c)

	if w.Code != 200 {
		t.Fatalf("DeliverBatch status = %d, want 200", w.Code)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, order := range store.data.Orders[:2] {
		status, _ := order["status"].(string)
		if status != "delivered" {
			t.Fatalf("order %v status = %q, want delivered", order["order_id"], status)
		}
	}
	status, _ := store.data.Orders[2]["status"].(string)
	if status != "assigned" {
		t.Fatalf("third order status = %q, want assigned", status)
	}
}
