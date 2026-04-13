package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPlatformDemoTriggerDisruptionRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupPlatformRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/demo/trigger-disruption", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing auth, got %d", w.Code)
	}
}

func TestPlatformWebhookAcceptsValidWebhookKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("INDEL_PLATFORM_WEBHOOK_KEY", "test-webhook-key")
	defer os.Unsetenv("INDEL_PLATFORM_WEBHOOK_KEY")

	r := gin.New()
	SetupPlatformRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/webhooks/external-signal", nil)
	req.Header.Set("X-Platform-Webhook-Key", "test-webhook-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Auth middleware should pass, handler returns 400 for missing payload.
	if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
		t.Fatalf("expected webhook key to bypass auth middleware, got %d", w.Code)
	}
}

func TestCoreInternalGenerateClaimsRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupCoreRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/claims/generate-for-disruption/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing auth, got %d", w.Code)
	}
}
