package platform

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMockPlatformDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test_platform.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Exec(`CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		worker_id INTEGER,
		zone_id INTEGER,
		order_value REAL,
		status TEXT,
		pickup_area TEXT,
		drop_area TEXT,
		distance_km REAL,
		vehicle_type TEXT,
		vehicle_capacity INTEGER,
		allowed_zones TEXT,
		updated_at DATETIME
	)`)
	return db
}

func setupPlatformRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	SetDB(db)
	if platformDB == nil {
		panic("platformDB is nil after SetDB in setupPlatformRouter")
	}
	r.POST("/api/platform/order/assign", OrderAssignedWebhook)
	return r
}

func TestOrderAssignedWebhook_WithVehicleFields(t *testing.T) {
	db := setupMockPlatformDB()
	r := setupPlatformRouter(db)
	if platformDB == nil {
		t.Fatalf("platformDB is nil after SetDB in test")
	}

	order := map[string]interface{}{
		"worker_id":        1,
		"zone_id":          2,
		"order_value":      100.0,
		"pickup_area":      "Whitefield",
		"drop_area":        "Koramangala",
		"distance_km":      12.5,
		"vehicle_type":     "van",
		"vehicle_capacity": 30,
		"allowed_zones":    "1,2,3",
	}
	body, _ := json.Marshal(order)
	req, _ := http.NewRequest("POST", "/api/platform/order/assign", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %v", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "assigned" {
		t.Errorf("Expected status 'assigned', got %v", resp["status"])
	}
}
