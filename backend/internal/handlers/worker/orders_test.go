package worker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAuthedOrderContext(orderID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderID+"/deliver", nil)
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	c.Request = req
	c.Params = gin.Params{{Key: "order_id", Value: orderID}}
	return c, w
}

func TestDeliverOrderIsIdempotentAfterFirstDelivery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	c1, w1 := newAuthedOrderContext("ord-001")
	DeliverOrder(c1)
	if w1.Code != 200 {
		t.Fatalf("first DeliverOrder status = %d, want 200", w1.Code)
	}

	store.mu.RLock()
	firstActual, _ := store.data.Earnings["this_week_actual"].(int)
	firstCompleted, _ := store.data.WorkerProfiles["worker-001"]["orders_completed"].(int)
	firstNotifications := len(store.data.Notifications)
	store.mu.RUnlock()

	if firstActual != 3198 {
		t.Fatalf("this_week_actual after first delivery = %d, want 3198", firstActual)
	}
	if firstCompleted != 1 {
		t.Fatalf("orders_completed after first delivery = %d, want 1", firstCompleted)
	}

	c2, w2 := newAuthedOrderContext("ord-001")
	DeliverOrder(c2)
	if w2.Code != 200 {
		t.Fatalf("second DeliverOrder status = %d, want 200", w2.Code)
	}

	store.mu.RLock()
	secondActual, _ := store.data.Earnings["this_week_actual"].(int)
	secondCompleted, _ := store.data.WorkerProfiles["worker-001"]["orders_completed"].(int)
	secondNotifications := len(store.data.Notifications)
	store.mu.RUnlock()

	if secondActual != firstActual {
		t.Fatalf("this_week_actual changed on repeated delivery: got %d, first %d", secondActual, firstActual)
	}
	if secondCompleted != firstCompleted {
		t.Fatalf("orders_completed changed on repeated delivery: got %d, first %d", secondCompleted, firstCompleted)
	}
	if secondNotifications != firstNotifications {
		t.Fatalf("notifications changed on repeated delivery: got %d, first %d", secondNotifications, firstNotifications)
	}
}

func TestGetAssignedOrdersRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/orders/assigned", nil)

	GetAssignedOrders(c)

	if w.Code != 401 {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
