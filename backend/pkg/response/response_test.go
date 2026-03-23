package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSuccessResponseIncludesMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "req-123")

	Success(c, 200, gin.H{"ok": true})

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	meta, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta missing or wrong type: %T", body["meta"])
	}
	if meta["request_id"] != "req-123" {
		t.Fatalf("request_id = %v, want req-123", meta["request_id"])
	}
	ts, ok := meta["timestamp"].(string)
	if !ok || ts == "" {
		t.Fatalf("timestamp = %v, want non-empty RFC3339 string", meta["timestamp"])
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("timestamp parse error: %v", err)
	}
}

func TestErrorResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, 400, "bad_request", "invalid payload")

	if w.Code != 400 {
		t.Fatalf("status = %d, want 400", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	errorBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("error field missing or wrong type: %T", body["error"])
	}
	if errorBody["code"] != "bad_request" {
		t.Fatalf("code = %v, want bad_request", errorBody["code"])
	}
	if errorBody["message"] != "invalid payload" {
		t.Fatalf("message = %v, want invalid payload", errorBody["message"])
	}
}
