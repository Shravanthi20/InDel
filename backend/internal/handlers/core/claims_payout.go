package core

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type payoutRequest struct {
	Amount float64 `json:"amount"`
}

// QueueClaimPayout queues payout for a claim via internal API.
// POST /internal/v1/claims/:claim_id/payout
func QueueClaimPayout(c *gin.Context) {
	if !hasDB() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db_unavailable"})
		return
	}

	claimIDParam := strings.TrimSpace(c.Param("claim_id"))
	claimIDParam = strings.TrimPrefix(claimIDParam, "clm-")
	claimID, err := strconv.ParseUint(claimIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_claim_id"})
		return
	}

	var req payoutRequest
	_ = c.ShouldBindJSON(&req)

	type claimRow struct {
		WorkerID    uint    `gorm:"column:worker_id"`
		ClaimAmount float64 `gorm:"column:claim_amount"`
	}
	var claim claimRow
	_ = coreDB.Raw("SELECT worker_id, claim_amount FROM claims WHERE id = ?", claimID).Scan(&claim).Error
	if claim.WorkerID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "claim_not_found"})
		return
	}

	amount := req.Amount
	if amount <= 0 {
		amount = claim.ClaimAmount
	}

	// Queue payout row (idempotent by unique claim_id).
	_ = coreDB.Exec(`
		INSERT INTO payouts (claim_id, worker_id, amount, status, razorpay_status)
		VALUES (?, ?, ?, 'queued', 'queued')
		ON CONFLICT (claim_id)
		DO UPDATE SET amount = EXCLUDED.amount, status = 'queued', razorpay_status = 'queued', updated_at = CURRENT_TIMESTAMP
	`, claimID, claim.WorkerID, amount).Error

	_ = coreDB.Exec("UPDATE claims SET status = 'queued_for_payout', updated_at = CURRENT_TIMESTAMP WHERE id = ?", claimID).Error

	// Simulate enqueue audit event for payout worker.
	_ = coreDB.Exec(`
		INSERT INTO kafka_event_logs (topic, event_type, payload_json)
		VALUES ('indel.payouts.queued', 'claim_payout_queued', jsonb_build_object('claim_id', ?, 'worker_id', ?, 'amount', ?))
	`, claimID, claim.WorkerID, amount).Error

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "payout_queued",
		"claim_id":  fmt.Sprintf("clm-%03d", claimID),
		"worker_id": claim.WorkerID,
		"amount":    int(amount),
		"topic":     "indel.payouts.queued",
	})
}
