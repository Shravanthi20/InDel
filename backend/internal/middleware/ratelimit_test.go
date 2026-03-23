package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddlewareBlocksBurstFromSameIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetLimiters()

	r := gin.New()
	r.Use(RateLimitMiddleware(1))
	r.GET("/limited", func(c *gin.Context) { c.Status(200) })

	req1 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req1.RemoteAddr = "1.2.3.4:1111"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Fatalf("first status = %d, want 200", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req2.RemoteAddr = "1.2.3.4:2222"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 429 {
		t.Fatalf("second status = %d, want 429", w2.Code)
	}
}

func TestRateLimitMiddlewareHandlesRemoteAddrWithoutPort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetLimiters()

	r := gin.New()
	r.Use(RateLimitMiddleware(1))
	r.GET("/limited", func(c *gin.Context) { c.Status(200) })

	req1 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req1.RemoteAddr = "5.6.7.8"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Fatalf("first status = %d, want 200", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req2.RemoteAddr = "5.6.7.8"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 429 {
		t.Fatalf("second status = %d, want 429", w2.Code)
	}
}
