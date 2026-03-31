package worker

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const publisherLeaseWindow = 5 * time.Minute

type publisherLeaseState struct {
	mu        sync.RWMutex
	SessionID string
	ExpiresAt time.Time
	LastAckAt time.Time
}

var demoPublisherLease = &publisherLeaseState{}

func newSessionID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "session-fallback"
	}
	return hex.EncodeToString(b)
}

func (p *publisherLeaseState) activate() (sessionID string, expiresAt time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.SessionID = newSessionID()
	p.LastAckAt = time.Now().UTC()
	p.ExpiresAt = p.LastAckAt.Add(publisherLeaseWindow)
	return p.SessionID, p.ExpiresAt
}

func (p *publisherLeaseState) extend(sessionID string) (expiresAt time.Time, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UTC()
	if p.SessionID == "" || now.After(p.ExpiresAt) {
		return time.Time{}, false
	}
	if strings.TrimSpace(sessionID) != "" && sessionID != p.SessionID {
		return time.Time{}, false
	}

	p.LastAckAt = now
	p.ExpiresAt = now.Add(publisherLeaseWindow)
	return p.ExpiresAt, true
}

func (p *publisherLeaseState) status() (active bool, sessionID string, expiresAt time.Time, remainingSeconds int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now().UTC()
	if p.SessionID == "" || now.After(p.ExpiresAt) {
		return false, "", time.Time{}, 0
	}
	remaining := int(time.Until(p.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return true, p.SessionID, p.ExpiresAt, remaining
}

func enforcePublisherControlKey(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("PUBLISHER_CONTROL_KEY"))
	if expected == "" {
		return true
	}
	provided := strings.TrimSpace(c.GetHeader("X-Publisher-Key"))
	if provided == expected {
		return true
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_publisher_control_key"})
	return false
}

type leaseAckRequest struct {
	SessionID string `json:"session_id"`
}

// InitiateOrderPublisher starts a 5-minute publishing lease.
func InitiateOrderPublisher(c *gin.Context) {
	if !enforcePublisherControlKey(c) {
		return
	}
	sessionID, expiresAt := demoPublisherLease.activate()
	c.JSON(http.StatusOK, gin.H{
		"message":             "publisher_initiated",
		"active":              true,
		"session_id":          sessionID,
		"lease_window_sec":    int(publisherLeaseWindow.Seconds()),
		"expires_at":          expiresAt.Format(time.RFC3339),
		"server_time_utc":     time.Now().UTC().Format(time.RFC3339),
		"publish_for_minutes": 5,
	})
}

// AckOrderPublisher extends the active lease by another 5 minutes.
func AckOrderPublisher(c *gin.Context) {
	if !enforcePublisherControlKey(c) {
		return
	}
	var req leaseAckRequest
	_ = c.ShouldBindJSON(&req)

	expiresAt, ok := demoPublisherLease.extend(req.SessionID)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "publisher_not_active_or_session_mismatch",
			"message": "call initiate first, then ack using current session_id",
		})
		return
	}

	active, sessionID, _, remaining := demoPublisherLease.status()
	c.JSON(http.StatusOK, gin.H{
		"message":          "publisher_acknowledged",
		"active":           active,
		"session_id":       sessionID,
		"lease_window_sec": int(publisherLeaseWindow.Seconds()),
		"remaining_sec":    remaining,
		"expires_at":       expiresAt.Format(time.RFC3339),
		"server_time_utc":  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetOrderPublisherStatus returns whether the publisher should be sending data.
func GetOrderPublisherStatus(c *gin.Context) {
	if !enforcePublisherControlKey(c) {
		return
	}
	active, sessionID, expiresAt, remaining := demoPublisherLease.status()
	resp := gin.H{
		"active":           active,
		"session_id":       sessionID,
		"remaining_sec":    remaining,
		"lease_window_sec": int(publisherLeaseWindow.Seconds()),
		"server_time_utc":  time.Now().UTC().Format(time.RFC3339),
	}
	if active {
		resp["expires_at"] = expiresAt.Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, resp)
}
