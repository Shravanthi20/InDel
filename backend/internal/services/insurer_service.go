package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

type InsurerService struct {
	DB            *gorm.DB
	KafkaProducer *kafka.Producer
}

func NewInsurerService(db *gorm.DB, kp *kafka.Producer) *InsurerService {
	return &InsurerService{DB: db, KafkaProducer: kp}
}

// GetOverview returns KPI overview
func (s *InsurerService) GetOverview() (*models.InsurerOverview, string, error) {
	if s.DB == nil {
		return &models.InsurerOverview{
			ActiveWorkers:      500,
			PendingClaims:      10,
			ApprovedClaims:     120,
			LossRatio:          0.45,
			ReserveUtilization: 0.45,
			Reserve:            1260,
		}, "healthy", nil
	}

	var activeWorkers int64
	var pendingClaims int64
	var approvedClaims int64
	var premiums float64
	var payouts float64

	_ = s.DB.Raw("SELECT COUNT(DISTINCT worker_id) FROM policies WHERE status = 'active'").Scan(&activeWorkers).Error
	_ = s.DB.Raw("SELECT COUNT(*) FROM claims WHERE status IN ('pending', 'manual_review')").Scan(&pendingClaims).Error
	_ = s.DB.Raw("SELECT COUNT(*) FROM claims WHERE status IN ('approved', 'processed', 'paid')").Scan(&approvedClaims).Error
	_ = s.DB.Raw("SELECT COALESCE(SUM(amount), 0) FROM premium_payments WHERE status IN ('completed', 'captured', 'processed')").Scan(&premiums).Error
	_ = s.DB.Raw("SELECT COALESCE(SUM(amount), 0) FROM payouts WHERE status IN ('processed', 'credited', 'completed')").Scan(&payouts).Error

	lossRatio := 0.0
	reserveUtilization := 0.0
	if premiums > 0 {
		lossRatio = payouts / premiums
		reserveUtilization = payouts / premiums
	}

	poolHealth := "healthy"
	if lossRatio > 0.8 {
		poolHealth = "watch"
	}
	if lossRatio > 1.0 {
		poolHealth = "critical"
	}

	return &models.InsurerOverview{
		ActiveWorkers:      float64(activeWorkers),
		PendingClaims:      float64(pendingClaims),
		ApprovedClaims:     float64(approvedClaims),
		LossRatio:          lossRatio,
		ReserveUtilization: reserveUtilization,
		Reserve:            premiums - payouts,
	}, poolHealth, nil
}

// GetLossRatio returns aggregated claims vs premiums
func (s *InsurerService) GetLossRatio(zoneID string) ([]models.LossRatio, error) {
	if s.DB == nil {
		return []models.LossRatio{{City: "Chennai", ZoneName: "Tambaram", Premiums: 2200, Claims: 980, LossRatio: 0.445}}, nil
	}

	type r struct {
		City     string
		Zone     string
		Premiums float64
		Claims   float64
	}
	var rows []r

	query := `
		SELECT z.city,
			   z.name AS zone,
			   COALESCE(p.premiums, 0) AS premiums,
			   COALESCE(cl.claims, 0) AS claims
		FROM zones z
		LEFT JOIN (
			SELECT wp.zone_id, SUM(pp.amount) AS premiums
			FROM premium_payments pp
			JOIN worker_profiles wp ON wp.worker_id = pp.worker_id
			WHERE pp.status IN ('completed', 'captured', 'processed')
			GROUP BY wp.zone_id
		) p ON p.zone_id = z.id
		LEFT JOIN (
			SELECT d.zone_id, SUM(c.claim_amount) AS claims
			FROM claims c
			JOIN disruptions d ON d.id = c.disruption_id
			GROUP BY d.zone_id
		) cl ON cl.zone_id = z.id
		WHERE (z.name = ? OR ? = '')
	`
	_ = s.DB.Raw(query, zoneID, zoneID).Scan(&rows).Error

	results := make([]models.LossRatio, 0, len(rows))
	for _, row := range rows {
		lr := 0.0
		if row.Premiums > 0 {
			lr = row.Claims / row.Premiums
		}
		results = append(results, models.LossRatio{
			City:      row.City,
			ZoneName:  row.Zone,
			Premiums:  row.Premiums,
			Claims:    row.Claims,
			LossRatio: lr,
		})
	}
	return results, nil
}

// GetClaims paginates the claims table
func (s *InsurerService) GetClaims(status string, fraudVerdict string, offset int, limit int) ([]models.ClaimListItem, int64, error) {
	if s.DB == nil {
		return []models.ClaimListItem{}, 0, nil
	}

	type r struct {
		ClaimID      uint
		Status       string
		City         string
		Zone         string
		ClaimAmount  float64
		FraudVerdict string
		CreatedAt    string
	}
	var rows []r
	var total int64

	baseQuery := s.DB.Table("claims c").
		Select("c.id AS claim_id, c.status, z.city, z.name AS zone, c.claim_amount, COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict, CAST(c.created_at as text) AS created_at").
		Joins("JOIN disruptions d ON d.id = c.disruption_id").
		Joins("JOIN zones z ON z.id = d.zone_id")

	countQuery := s.DB.Table("claims c").
		Joins("JOIN disruptions d ON d.id = c.disruption_id").
		Joins("JOIN zones z ON z.id = d.zone_id")

	if status != "" {
		baseQuery = baseQuery.Where("c.status = ?", status)
		countQuery = countQuery.Where("c.status = ?", status)
	}
	if fraudVerdict != "" {
		if fraudVerdict == "pending" {
			baseQuery = baseQuery.Where("COALESCE(c.fraud_verdict, 'pending') = ?", fraudVerdict)
			countQuery = countQuery.Where("COALESCE(c.fraud_verdict, 'pending') = ?", fraudVerdict)
		} else {
			baseQuery = baseQuery.Where("c.fraud_verdict = ?", fraudVerdict)
			countQuery = countQuery.Where("c.fraud_verdict = ?", fraudVerdict)
		}
	}

	_ = countQuery.Count(&total)

	_ = baseQuery.Order("c.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	results := make([]models.ClaimListItem, 0, len(rows))
	for _, row := range rows {
		t, _ := time.Parse("2006-01-02 15:04:05.999999999-07:00", row.CreatedAt)
		results = append(results, models.ClaimListItem{
			ClaimID:      row.ClaimID,
			ZoneName:     row.Zone,
			Status:       row.Status,
			ClaimAmount:  row.ClaimAmount,
			FraudVerdict: row.FraudVerdict,
			CreatedAt:    t,
		})
	}
	return results, total, nil
}

// GetClaimDetail joins ML scores
func (s *InsurerService) GetClaimDetail(claimID string) (*models.ClaimDetail, error) {
	if s.DB == nil {
		return &models.ClaimDetail{
			ClaimID:           "clm_x1",
			WorkerID:          "wkr_x1",
			ZoneID:            "zone_tambaram_chennai",
			LossAmount:        740.25,
			RecommendedPayout: 518.18,
			Status:            "pending",
			FraudVerdict:      "review",
			FraudScore:        0.73,
			Factors: []models.FraudFactor{
				{Name: "gps_mismatch", Impact: 0.24},
			},
			CreatedAt: "2026-03-30T10:00:00Z",
		}, nil
	}

	type r struct {
		ClaimID      uint
		WorkerID     uint
		DisruptionID uint
		ZoneName     string
		City         string
		ClaimAmount  float64
		Status       string
		FraudVerdict string
		FraudScore   float64
		Factors      []byte
		CreatedAt    string
	}
	var row r
	err := s.DB.Table("claims c").
		Select("c.id AS claim_id, c.worker_id, c.disruption_id, z.name AS zone_name, z.city, c.claim_amount, c.status, COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict, COALESCE(cfs.isolation_forest_score, 0.0) AS fraud_score, cfs.rule_violations AS factors, CAST(c.created_at as text) AS created_at").
		Joins("JOIN disruptions d ON d.id = c.disruption_id").
		Joins("JOIN zones z ON z.id = d.zone_id").
		Joins("LEFT JOIN claim_fraud_scores cfs ON cfs.claim_id = c.id").
		Where("c.id = ?", claimID).
		First(&row).Error

	if err != nil {
		return nil, fmt.Errorf("claim not found")
	}

	var factors []models.FraudFactor
	if len(row.Factors) > 0 {
		_ = json.Unmarshal(row.Factors, &factors)
	} else {
		factors = []models.FraudFactor{}
	}

	return &models.ClaimDetail{
		ClaimID:           fmt.Sprintf("clm_%d", row.ClaimID),
		WorkerID:          fmt.Sprintf("wkr_%d", row.WorkerID),
		ZoneID:            fmt.Sprintf("zone_%s_%s", row.ZoneName, row.City),
		DisruptionID:      fmt.Sprintf("dis_%d", row.DisruptionID),
		LossAmount:        row.ClaimAmount,
		RecommendedPayout: row.ClaimAmount * 0.70,
		Status:            row.Status,
		FraudVerdict:      row.FraudVerdict,
		FraudScore:        row.FraudScore,
		Factors:           factors,
		CreatedAt:         row.CreatedAt,
	}, nil
}

// ReviewClaim processes manual decision and emits event
func (s *InsurerService) ReviewClaim(claimID string, req models.ClaimAction) error {
	if s.DB == nil {
		return nil
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Exec("UPDATE claims SET status = ?, fraud_verdict = ?, updated_at = ? WHERE id = ?", req.Status, req.FraudVerdict, time.Now(), claimID)
		if res.Error != nil {
			return res.Error
		}

		cid := 0
		if _, err := fmt.Sscanf(claimID, "%d", &cid); err != nil {
			return err
		}

		audit := models.ClaimAuditLog{
			ClaimID:   uint(cid),
			Action:    "review",
			Notes:     req.Notes,
			Reviewer:  "system_user",
			CreatedAt: time.Now(),
		}

		if err := tx.Create(&audit).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to review claim: %w", err)
	}

	// Emit Kafka event
	if s.KafkaProducer != nil {
		ev := map[string]interface{}{
			"event_type":    "claim.reviewed",
			"claim_id":      claimID,
			"status":        req.Status,
			"fraud_verdict": req.FraudVerdict,
			"timestamp":     time.Now().Format(time.RFC3339),
		}
		b, _ := json.Marshal(ev)
		_ = s.KafkaProducer.Publish(kafka.TopicClaimReviewed, claimID, b)
	}

	return nil
}

// GetFraudQueue list ML flagged claims
func (s *InsurerService) GetFraudQueue(offset, limit int) ([]models.FraudQueueItem, int64, error) {
	if s.DB == nil {
		return []models.FraudQueueItem{{ClaimID: 1, FraudVerdict: "pending"}}, 1, nil
	}
	type r struct {
		ClaimID      uint
		FinalVerdict string
		Violations   string
		CreatedAt    string
		Score        float64
	}
	var rows []r
	var total int64

	baseQuery := s.DB.Table("claims c").
		Select("c.id AS claim_id, COALESCE(cfs.final_verdict, 'pending') AS final_verdict, COALESCE(cfs.isolation_forest_score, 0.0) AS score, COALESCE(CAST(cfs.rule_violations as text), '[]') AS violations, CAST(c.created_at as text) AS created_at").
		Joins("LEFT JOIN claim_fraud_scores cfs ON cfs.claim_id = c.id").
		Where("COALESCE(cfs.final_verdict, 'pending') IN ('flagged', 'manual_review', 'pending')")

	_ = s.DB.Table("claims c").
		Joins("LEFT JOIN claim_fraud_scores cfs ON cfs.claim_id = c.id").
		Where("COALESCE(cfs.final_verdict, 'pending') IN ('flagged', 'manual_review', 'pending')").
		Count(&total).Error

	_ = baseQuery.Order("cfs.isolation_forest_score DESC, c.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	results := make([]models.FraudQueueItem, 0, len(rows))
	for _, row := range rows {
		results = append(results, models.FraudQueueItem{
			ClaimID:      row.ClaimID,
			Status:       "manual_review", // Contextual
			FraudVerdict: row.FinalVerdict,
			FraudScore:   row.Score,
			CreatedAt:    row.CreatedAt,
		})
	}

	return results, total, nil
}
