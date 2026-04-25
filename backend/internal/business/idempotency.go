package business

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/logging"
	"gorm.io/gorm"
)

// IdempotencyKey represents an idempotency key for operations
type IdempotencyKey struct {
	ID          uint      `gorm:"primaryKey"`
	Key         string    `gorm:"uniqueIndex;not null"`
	Operation   string    `gorm:"not null"` // Type of operation (e.g., "policy_create", "claim_submit")
	Status      string    `gorm:"not null"` // "processing", "completed", "failed"
	Result      string    `gorm:"type:text"` // JSON result of the operation
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// IdempotencyManager handles idempotency for business operations
type IdempotencyManager struct {
	db     *gorm.DB
	logger *logging.ContextLogger
}

// NewIdempotencyManager creates a new idempotency manager
func NewIdempotencyManager(db *gorm.DB) *IdempotencyManager {
	return &IdempotencyManager{
		db:     db,
		logger: logging.WithContext(map[string]interface{}{
			"service": "idempotency-manager",
		}),
	}
}

// CheckAndSet checks if an operation with the given key has been processed
// and sets it to processing status if not found
func (im *IdempotencyManager) CheckAndSet(idempotencyKey, operation string, ttl time.Duration) (*IdempotencyKey, bool, error) {
	logger := im.logger.WithContext(map[string]interface{}{
		"idempotency_key": idempotencyKey,
		"operation": operation,
	})

	// Clean up expired keys first
	if err := im.cleanupExpiredKeys(); err != nil {
		logger.Warn("Failed to cleanup expired keys", nil, map[string]interface{}{
			"error": err.Error(),
		})
	}

	var existing IdempotencyKey
	err := im.db.Where("key = ? AND operation = ?", idempotencyKey, operation).First(&existing).Error
	
	if err == nil {
		// Key exists, check status
		switch existing.Status {
		case "processing":
			logger.Warn("Operation is already being processed", nil)
			return &existing, false, fmt.Errorf("operation is already being processed")
		case "completed":
			logger.Debug("Operation was already completed", nil)
			return &existing, false, nil
		case "failed":
			logger.Debug("Operation previously failed, allowing retry", nil)
			// Delete failed entry and allow retry
			im.db.Delete(&existing)
		default:
			logger.Warn("Unknown operation status", nil, map[string]interface{}{
				"status": existing.Status,
			})
		}
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check idempotency key", err)
		return nil, false, fmt.Errorf("database error: %w", err)
	}

	// Create new idempotency key
	now := time.Now()
	newKey := IdempotencyKey{
		Key:       idempotencyKey,
		Operation: operation,
		Status:    "processing",
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}

	if err := im.db.Create(&newKey).Error; err != nil {
		logger.Error("Failed to create idempotency key", err)
		return nil, false, fmt.Errorf("failed to create idempotency key: %w", err)
	}

	logger.Debug("Created new idempotency key", nil)
	return &newKey, true, nil
}

// Complete marks an operation as completed with the given result
func (im *IdempotencyManager) Complete(idempotencyKey *IdempotencyKey, result interface{}) error {
	logger := im.logger.WithContext(map[string]interface{}{
		"idempotency_key": idempotencyKey.Key,
		"operation": idempotencyKey.Operation,
	})

	// Serialize result
	resultJSON := "{}"
	if result != nil {
		if jsonBytes, err := serializeResult(result); err == nil {
			resultJSON = string(jsonBytes)
		} else {
			logger.Error("Failed to serialize result", err)
		}
	}

	// Update status
	if err := im.db.Model(idempotencyKey).Updates(map[string]interface{}{
		"status":  "completed",
		"result":  resultJSON,
	}).Error; err != nil {
		logger.Error("Failed to mark operation as completed", err)
		return fmt.Errorf("failed to complete operation: %w", err)
	}

	logger.Debug("Operation marked as completed", nil)
	return nil
}

// Fail marks an operation as failed
func (im *IdempotencyManager) Fail(idempotencyKey *IdempotencyKey, err error) error {
	logger := im.logger.WithContext(map[string]interface{}{
		"idempotency_key": idempotencyKey.Key,
		"operation": idempotencyKey.Operation,
	})

	errorJSON := "{}"
	if err != nil {
		errorJSON = fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}

	// Update status
	if dbErr := im.db.Model(idempotencyKey).Updates(map[string]interface{}{
		"status": "failed",
		"result": errorJSON,
	}).Error; dbErr != nil {
		logger.Error("Failed to mark operation as failed", dbErr)
		return fmt.Errorf("failed to fail operation: %w", dbErr)
	}

	logger.Debug("Operation marked as failed", nil)
	return nil
}

// GetResult retrieves the result of a completed operation
func (im *IdempotencyManager) GetResult(idempotencyKey, operation string) (*IdempotencyKey, error) {
	var key IdempotencyKey
	err := im.db.Where("key = ? AND operation = ? AND status = ?", 
		idempotencyKey, operation, "completed").First(&key).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no completed operation found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &key, nil
}

// cleanupExpiredKeys removes expired idempotency keys
func (im *IdempotencyManager) cleanupExpiredKeys() error {
	result := im.db.Where("expires_at < ?", time.Now()).Delete(&IdempotencyKey{})
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		im.logger.Debug("Cleaned up expired idempotency keys", nil, map[string]interface{}{
			"cleaned_count": result.RowsAffected,
		})
	}
	
	return nil
}

// serializeResult converts a result to JSON
func serializeResult(result interface{}) ([]byte, error) {
	// This would typically use encoding/json, but keeping it simple for now
	return []byte(fmt.Sprintf(`{"result": "%v"}`, result)), nil
}

// GenerateIdempotencyKey generates a unique idempotency key for an operation
func GenerateIdempotencyKey(operation, userID, resourceID string, timestamp time.Time) string {
	data := fmt.Sprintf("%s:%s:%s:%d", operation, userID, resourceID, timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BusinessLogicValidator provides validation for business logic edge cases
type BusinessLogicValidator struct {
	db     *gorm.DB
	logger *logging.ContextLogger
}

// NewBusinessLogicValidator creates a new business logic validator
func NewBusinessLogicValidator(db *gorm.DB) *BusinessLogicValidator {
	return &BusinessLogicValidator{
		db:     db,
		logger: logging.WithContext(map[string]interface{}{
			"service": "business-logic-validator",
		}),
	}
}

// ValidatePolicyTransition validates that a policy state transition is valid
func (blv *BusinessLogicValidator) ValidatePolicyTransition(currentStatus, newStatus string, workerID uint) error {
	logger := blv.logger.WithContext(map[string]interface{}{
		"current_status": currentStatus,
		"new_status": newStatus,
		"worker_id": workerID,
	})

	// Define valid transitions
	validTransitions := map[string][]string{
		"locked":   {"active", "cancelled"},
		"active":   {"paused", "cancelled", "expired"},
		"paused":   {"active", "cancelled"},
		"cancelled": {}, // Terminal state
		"expired":   {},  // Terminal state
	}

	// Check if transition is valid
	if allowedStates, exists := validTransitions[currentStatus]; exists {
		for _, allowed := range allowedStates {
			if allowed == newStatus {
				logger.Debug("Policy transition is valid", nil)
				return nil
			}
		}
	}

	logger.Warn("Invalid policy transition attempted", nil)
	return fmt.Errorf("invalid policy transition from %s to %s", currentStatus, newStatus)
}

// ValidatePolicyLockIn ensures policy lock-in rules are followed
func (blv *BusinessLogicValidator) ValidatePolicyLockIn(workerID uint, lockInHours int) error {
	logger := blv.logger.WithContext(map[string]interface{}{
		"worker_id": workerID,
		"lock_in_hours": lockInHours,
	})

	// Check if worker already has an active policy
	var activePolicyCount int64
	if err := blv.db.Table("policies").Where("worker_id = ? AND status IN ('active', 'locked')", workerID).Count(&activePolicyCount).Error; err != nil {
		logger.Error("Failed to check for existing policies", err)
		return fmt.Errorf("database error: %w", err)
	}

	if activePolicyCount > 0 {
		logger.Warn("Worker already has an active policy", nil)
		return fmt.Errorf("worker already has an active policy")
	}

	// Validate lock-in hours
	if lockInHours <= 0 || lockInHours > 168 { // Max 1 week
		logger.Warn("Invalid lock-in hours", nil)
		return fmt.Errorf("lock-in hours must be between 1 and 168")
	}

	logger.Debug("Policy lock-in validation passed", nil)
	return nil
}

// ValidateClaimSubmission validates claim submission rules
func (blv *BusinessLogicValidator) ValidateClaimSubmission(workerID uint, disruptionID uint) error {
	logger := blv.logger.WithContext(map[string]interface{}{
		"worker_id": workerID,
		"disruption_id": disruptionID,
	})

	// Check if worker has an active policy
	var activePolicyCount int64
	if err := blv.db.Table("policies").Where("worker_id = ? AND status = 'active'", workerID).Count(&activePolicyCount).Error; err != nil {
		logger.Error("Failed to check for active policy", err)
		return fmt.Errorf("database error: %w", err)
	}

	if activePolicyCount == 0 {
		logger.Warn("Worker has no active policy", nil)
		return fmt.Errorf("worker must have an active policy to submit claims")
	}

	// Check if disruption exists and is confirmed
	var disruption struct {
		Status string `gorm:"column:status"`
		ZoneID uint   `gorm:"column:zone_id"`
	}
	if err := blv.db.Table("disruptions").Where("id = ?", disruptionID).First(&disruption).Error; err != nil {
		logger.Error("Failed to check disruption", err)
		return fmt.Errorf("disruption not found: %w", err)
	}

	if disruption.Status != "confirmed" && disruption.Status != "active" {
		logger.Warn("Disruption is not active", nil)
		return fmt.Errorf("disruption must be confirmed or active")
	}

	// Check if worker is in the affected zone
	var workerZone struct {
		ZoneID uint `gorm:"column:zone_id"`
	}
	if err := blv.db.Table("worker_profiles").Where("worker_id = ?", workerID).First(&workerZone).Error; err != nil {
		logger.Error("Failed to check worker zone", err)
		return fmt.Errorf("worker not found: %w", err)
	}

	if workerZone.ZoneID != disruption.ZoneID {
		logger.Warn("Worker not in affected zone", nil)
		return fmt.Errorf("worker must be in the affected zone to submit claims")
	}

	// Check for existing claim for this disruption
	var existingClaimCount int64
	if err := blv.db.Table("claims").Where("worker_id = ? AND disruption_id = ?", workerID, disruptionID).Count(&existingClaimCount).Error; err != nil {
		logger.Error("Failed to check for existing claims", err)
		return fmt.Errorf("database error: %w", err)
	}

	if existingClaimCount > 0 {
		logger.Warn("Claim already exists for this disruption", nil)
		return fmt.Errorf("claim already exists for this disruption")
	}

	logger.Debug("Claim submission validation passed", nil)
	return nil
}

// ValidateDisruptionCreation validates disruption creation rules
func (blv *BusinessLogicValidator) ValidateDisruptionCreation(zoneID uint, disruptionType string) error {
	logger := blv.logger.WithContext(map[string]interface{}{
		"zone_id": zoneID,
		"disruption_type": disruptionType,
	})

	// Check if zone exists
	var zoneCount int64
	if err := blv.db.Table("zones").Where("id = ?", zoneID).Count(&zoneCount).Error; err != nil {
		logger.Error("Failed to check zone", err)
		return fmt.Errorf("database error: %w", err)
	}

	if zoneCount == 0 {
		logger.Warn("Zone does not exist", nil)
		return fmt.Errorf("zone does not exist")
	}

	// Validate disruption type
	validTypes := []string{"weather", "demand_drop", "infrastructure", "accident", "other"}
	typeValid := false
	for _, validType := range validTypes {
		if disruptionType == validType {
			typeValid = true
			break
		}
	}

	if !typeValid {
		logger.Warn("Invalid disruption type", nil)
		return fmt.Errorf("invalid disruption type: %s", disruptionType)
	}

	// Check for recent disruption in the same zone (to prevent duplicates)
	var recentDisruptionCount int64
	recentThreshold := time.Now().Add(-1 * time.Hour) // 1 hour ago
	if err := blv.db.Table("disruptions").Where("zone_id = ? AND created_at > ? AND status IN ('active', 'confirmed')", 
		zoneID, recentThreshold).Count(&recentDisruptionCount).Error; err != nil {
		logger.Error("Failed to check for recent disruptions", err)
		return fmt.Errorf("database error: %w", err)
	}

	if recentDisruptionCount > 0 {
		logger.Warn("Recent disruption already exists in zone", nil)
		return fmt.Errorf("recent disruption already exists in this zone")
	}

	logger.Debug("Disruption creation validation passed", nil)
	return nil
}

// ValidatePremiumCalculation validates premium calculation inputs
func (blv *BusinessLogicValidator) ValidatePremiumCalculation(workerID uint, requestedPremium float64) error {
	logger := blv.logger.WithContext(map[string]interface{}{
		"worker_id": workerID,
		"requested_premium": requestedPremium,
	})

	// Validate premium amount
	if requestedPremium <= 0 {
		logger.Warn("Invalid premium amount", nil)
		return fmt.Errorf("premium must be positive")
	}

	if requestedPremium > 10000 { // Max premium limit
		logger.Warn("Premium amount exceeds limit", nil)
		return fmt.Errorf("premium amount exceeds maximum limit")
	}

	// Check worker exists
	var workerCount int64
	if err := blv.db.Table("worker_profiles").Where("worker_id = ?", workerID).Count(&workerCount).Error; err != nil {
		logger.Error("Failed to check worker", err)
		return fmt.Errorf("database error: %w", err)
	}

	if workerCount == 0 {
		logger.Warn("Worker does not exist", nil)
		return fmt.Errorf("worker does not exist")
	}

	logger.Debug("Premium calculation validation passed", nil)
	return nil
}
