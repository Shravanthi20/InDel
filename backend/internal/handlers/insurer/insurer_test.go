package insurer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMockDB creates an in-memory SQLite map with the required schema.
func setupMockDB() *gorm.DB {
	dbName := fmt.Sprintf("file:testdb%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Create tables using manual queries representing the schema used in queries
	db.Exec(`
		CREATE TABLE policies (worker_id INTEGER, status TEXT);
		CREATE TABLE claims (id INTEGER PRIMARY KEY, disruption_id INTEGER, worker_id INTEGER, claim_amount REAL, status TEXT, fraud_verdict TEXT, created_at DATETIME, updated_at DATETIME);
		CREATE TABLE premium_payments (amount REAL, status TEXT, worker_id INTEGER);
		CREATE TABLE payouts (amount REAL, status TEXT);
		CREATE TABLE zones (id INTEGER PRIMARY KEY, city TEXT, name TEXT);
		CREATE TABLE disruptions (id INTEGER PRIMARY KEY, zone_id INTEGER);
		CREATE TABLE worker_profiles (worker_id INTEGER, zone_id INTEGER);
		CREATE TABLE claim_fraud_scores (claim_id INTEGER PRIMARY KEY, score REAL, final_verdict TEXT, rule_violations TEXT, created_at DATETIME, updated_at DATETIME);
		CREATE TABLE claim_audit_logs (id INTEGER PRIMARY KEY, claim_id INTEGER, action TEXT, notes TEXT, reviewer TEXT, created_at DATETIME);
	`)

	return db
}

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	svc := services.NewInsurerService(db, nil)
	h := NewInsurerHandler(svc)
	r.GET("/api/v1/insurer/overview", h.GetOverview)
	r.GET("/api/v1/insurer/loss-ratio", h.GetLossRatio)
	r.GET("/api/v1/insurer/claims", h.GetClaims)
	r.GET("/api/v1/insurer/claims/fraud-queue", h.GetFraudQueue)
	r.GET("/api/v1/insurer/claims/:id", h.GetClaimDetail)
	r.POST("/api/v1/insurer/claims/:id/review", h.ReviewClaim)
	return r
}

func TestOverviewAggregation(t *testing.T) {
	db := setupMockDB()
	r := setupRouter(db)

	// Insert mock data
	db.Exec("INSERT INTO policies (worker_id, status) VALUES (1, 'active'), (2, 'active'), (3, 'inactive')")
	db.Exec("INSERT INTO claims (status) VALUES ('pending'), ('manual_review'), ('approved')")
	db.Exec("INSERT INTO premium_payments (amount, status) VALUES (1000, 'completed'), (500, 'processed')")
	db.Exec("INSERT INTO payouts (amount, status) VALUES (300, 'processed')")

	req, _ := http.NewRequest("GET", "/api/v1/insurer/overview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	data := response.Data.(map[string]interface{})
	if data["active_workers"].(float64) != 2 {
		t.Errorf("Expected 2 active workers, got %v", data["active_workers"])
	}
	if data["pending_claims"].(float64) != 2 {
		t.Errorf("Expected 2 pending claims, got %v", data["pending_claims"])
	}
	if data["approved_claims"].(float64) != 1 {
		t.Errorf("Expected 1 approved claim, got %v", data["approved_claims"])
	}
	if data["reserve"].(float64) != 1200 {
		t.Errorf("Expected reserve 1200, got %v", data["reserve"])
	}
	if data["loss_ratio"].(float64) != 0.2 { // 300 / 1500
		t.Errorf("Expected loss ratio 0.2, got %v", data["loss_ratio"])
	}
}

func TestLossRatioEdgeCases(t *testing.T) {
	db := setupMockDB()
	r := setupRouter(db)

	// Zero premiums, nonzero claims
	db.Exec("INSERT INTO zones (id, city, name) VALUES (1, 'Chennai', 'Tambaram')")
	db.Exec("INSERT INTO disruptions (id, zone_id) VALUES (1, 1)")
	db.Exec("INSERT INTO claims (disruption_id, claim_amount) VALUES (1, 500)")

	req, _ := http.NewRequest("GET", "/api/v1/insurer/loss-ratio", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", w.Code)
	}

	var response SuccessResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	data := response.Data.([]interface{})
	zone := data[0].(map[string]interface{})

	if zone["loss_ratio"].(float64) != 0 {
		t.Errorf("Expected loss ratio to handle 0 premiums gracefully, got %v", zone["loss_ratio"])
	}
}

func TestClaimsPaginationAndFilters(t *testing.T) {
	db := setupMockDB()
	r := setupRouter(db)

	db.Exec("INSERT INTO zones (id, city, name) VALUES (1, 'A', 'B')")
	db.Exec("INSERT INTO disruptions (id, zone_id) VALUES (1, 1)")
	for i := 1; i <= 25; i++ {
		status := "pending"
		verdict := "pending"
		if i%2 == 0 {
			status = "approved"
		}
		if i == 5 {
			verdict = "flagged"
		}
		db.Exec("INSERT INTO claims (id, disruption_id, status, fraud_verdict, created_at) VALUES (?, 1, ?, ?, ?)", i, status, verdict, time.Now())
	}

	req, _ := http.NewRequest("GET", "/api/v1/insurer/claims?limit=10&page=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response PaginatedResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	data := response.Data.([]interface{})

	if len(data) != 10 {
		t.Errorf("Expected 10 items per page, got %v", len(data))
	}
	if response.Pagination.Total != 25 {
		t.Errorf("Expected total 25, got %v", response.Pagination.Total)
	}
	if !response.Pagination.HasNext {
		t.Error("Expected has_next to be true")
	}

	// Filter test
	req2, _ := http.NewRequest("GET", "/api/v1/insurer/claims?fraud_verdict=flagged", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp2 PaginatedResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
	data2 := resp2.Data.([]interface{})
	if len(data2) != 1 {
		t.Errorf("Expected 1 filtered item, got %v", len(data2))
	}
}

func TestReviewClaimIdempotencyAndAudit(t *testing.T) {
	db := setupMockDB()
	r := setupRouter(db)

	db.Exec("INSERT INTO claims (id, status, fraud_verdict) VALUES (1, 'pending', 'review')")

	payload := []byte(`{"status": "denied", "fraud_verdict": "fraud", "notes": "GPS anomalies"}`)
	
	// First call
	req, _ := http.NewRequest("POST", "/api/v1/insurer/claims/1/review", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %v", w.Code)
	}

	// Second call
	req2, _ := http.NewRequest("POST", "/api/v1/insurer/claims/1/review", bytes.NewBuffer(payload))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected 200 on repeat call, got %v", w2.Code)
	}

	var claimStatus string
	db.Raw("SELECT status FROM claims WHERE id = 1").Scan(&claimStatus)
	if claimStatus != "denied" {
		t.Errorf("Expected status denied, got %s", claimStatus)
	}

	var auditCount int64
	db.Raw("SELECT COUNT(*) FROM claim_audit_logs WHERE claim_id = 1").Scan(&auditCount)
	if auditCount != 2 {
		t.Errorf("Expected 2 audit log entries, got %d", auditCount)
	}
}

func TestFraudQueueSelection(t *testing.T) {
	db := setupMockDB()
	r := setupRouter(db)

	db.Exec("INSERT INTO claims (id) VALUES (1), (2), (3)")
	db.Exec("INSERT INTO claim_fraud_scores (claim_id, final_verdict) VALUES (1, 'flagged'), (2, 'clear'), (3, 'manual_review')")

	req, _ := http.NewRequest("GET", "/api/v1/insurer/claims/fraud-queue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response PaginatedResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	data := response.Data.([]interface{})

	if len(data) != 2 {
		t.Errorf("Expected 2 items in fraud queue, got %d", len(data))
	}
}
