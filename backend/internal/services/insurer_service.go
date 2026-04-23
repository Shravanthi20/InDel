package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

type InsurerService struct {
	DB            *gorm.DB
	KafkaProducer *kafka.Producer
}

type MaintenanceCheckItem struct {
	ID           uint      `json:"id"`
	ClaimID      uint      `json:"claim_id"`
	WorkerID     uint      `json:"worker_id"`
	ZoneName     string    `json:"zone_name"`
	City         string    `json:"city"`
	Status       string    `json:"status"`
	FraudVerdict string    `json:"fraud_verdict"`
	ClaimAmount  float64   `json:"claim_amount"`
	InitiatedAt  time.Time `json:"initiated_at"`
	ResponseAt   *string   `json:"response_at,omitempty"`
	Findings     string    `json:"findings"`
}

type ZoneMoneyExchange struct {
	ZoneID            uint    `json:"zone_id"`
	ZoneName          string  `json:"zone_name"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Level             string  `json:"level"`
	SubscribedWorkers int64   `json:"subscribed_workers"`
	ClaimsCount       int64   `json:"claims_count"`
	PremiumsCollected float64 `json:"premiums_collected"`
	ClaimsAmount      float64 `json:"claims_amount"`
	PayoutsProcessed  float64 `json:"payouts_processed"`
	NetFlow           float64 `json:"net_flow"`
}

type MoneyExchangeSummary struct {
	PremiumPool      float64             `json:"premium_pool"`
	TotalSubscribed  int64               `json:"total_subscribed"`
	TotalClaims      int64               `json:"total_claims"`
	TotalClaimAmount float64             `json:"total_claim_amount"`
	TotalPayouts     float64             `json:"total_payouts"`
	NetPool          float64             `json:"net_pool"`
	PendingPayouts   int64               `json:"pending_payouts"`
	ZoneBreakdown    []ZoneMoneyExchange `json:"zone_breakdown"`
}

type UserPlanStatus struct {
	UserID         uint                 `json:"id"`
	Name           string               `json:"name"`
	Phone          string               `json:"phone"`
	Zone           string               `json:"zone"`
	Status         string               `json:"status"`
	PolicyID       *uint                `json:"policy_id,omitempty"`
	PlanID         string               `json:"plan_id,omitempty"`
	StartedAt      string               `json:"started_at,omitempty"`
	UpdatedAt      string               `json:"updated_at,omitempty"`
	WeeklyPremium  float64              `json:"weekly_premium"`
	MaxPayout      float64              `json:"max_payout" gorm:"-"`
	Explainability []PremiumExplainItem `json:"explainability,omitempty" gorm:"-"`
}

func NewInsurerService(db *gorm.DB, kp *kafka.Producer) *InsurerService {
	return &InsurerService{DB: db, KafkaProducer: kp}
}

func (s *InsurerService) ListUserPlanStatuses() ([]UserPlanStatus, error) {
	if s.DB == nil {
		return []UserPlanStatus{}, nil
	}

	rows := make([]UserPlanStatus, 0)
	err := s.DB.Raw(`
		SELECT
			u.id AS id,
			COALESCE(NULLIF(TRIM(wp.name), ''), u.phone, 'User') AS name,
			COALESCE(u.phone, '') AS phone,
			COALESCE(z.name, 'Unknown') AS zone,
			CASE WHEN ap.user_id IS NOT NULL THEN 'active' ELSE 'inactive' END AS status,
			lp.id AS policy_id,
			COALESCE(lp.premium_amount, 0) AS weekly_premium,
			COALESCE(lp.plan_id, '') AS plan_id,
			COALESCE(CAST(ap.started_at AS text), '') AS started_at,
			COALESCE(CAST(COALESCE(ap.updated_at, lp.updated_at, lp.created_at) AS text), '') AS updated_at
		FROM users u
		LEFT JOIN worker_profiles wp ON wp.worker_id = u.id
		LEFT JOIN zones z ON z.id = wp.zone_id
		LEFT JOIN active_policies ap ON ap.user_id = u.id
		LEFT JOIN policies lp ON lp.id = (
			SELECT p2.id
			FROM policies p2
			WHERE p2.worker_id = u.id
			ORDER BY p2.id DESC
			LIMIT 1
		)
		ORDER BY u.id ASC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Compute fresh premiums for inactive workers, or enrich active ones
	for i := range rows {
		if rows[i].Status == "active" && rows[i].WeeklyPremium > 0 {
			// SYNC: Use the actual committed premium from the database (e.g. ₹49 for Rose)
			// Still compute explainability to show 'Why this premium?'
			quote, _ := QuotePremium(s.DB, rows[i].UserID, time.Now().UTC())
			if quote != nil {
				rows[i].Explainability = quote.Explainability
				rows[i].MaxPayout = 500 + (quote.RiskScore * 400)
			} else {
				rows[i].MaxPayout = 800
			}
		} else {
			// For unprotected workers, compute a live quote
			quote, _ := QuotePremium(s.DB, rows[i].UserID, time.Now().UTC())
			if quote != nil {
				rows[i].WeeklyPremium = quote.WeeklyPremiumINR
				rows[i].Explainability = quote.Explainability
				rows[i].MaxPayout = 500 + (quote.RiskScore * 400)
			} else {
				// Fallback defaults
				rows[i].WeeklyPremium = 22
				rows[i].MaxPayout = 800
			}
		}
	}

	return rows, nil
}

func (s *InsurerService) GetUserPlanStatus(userID uint) (*UserPlanStatus, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	rows := make([]UserPlanStatus, 0, 1)
	err := s.DB.Raw(`
		SELECT
			u.id AS id,
			COALESCE(NULLIF(TRIM(wp.name), ''), u.phone, 'User') AS name,
			COALESCE(u.phone, '') AS phone,
			COALESCE(z.name, 'Unknown') AS zone,
			CASE WHEN ap.user_id IS NOT NULL THEN 'active' ELSE 'inactive' END AS status,
			lp.id AS policy_id,
			COALESCE(lp.plan_id, '') AS plan_id,
			COALESCE(CAST(ap.started_at AS text), '') AS started_at,
			COALESCE(CAST(COALESCE(ap.updated_at, lp.updated_at, lp.created_at) AS text), '') AS updated_at
		FROM users u
		LEFT JOIN worker_profiles wp ON wp.worker_id = u.id
		LEFT JOIN zones z ON z.id = wp.zone_id
		LEFT JOIN active_policies ap ON ap.user_id = u.id
		LEFT JOIN policies lp ON lp.id = (
			SELECT p2.id
			FROM policies p2
			WHERE p2.worker_id = u.id
			ORDER BY p2.id DESC
			LIMIT 1
		)
		WHERE u.id = ?
		LIMIT 1
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &rows[0], nil
}

func (s *InsurerService) StartUserPlan(userID uint) (*UserPlanStatus, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		type userContext struct {
			ID   uint
			Zone string
		}

		var ctx userContext
		if err := tx.Raw(`
			SELECT u.id, COALESCE(z.name, 'Unknown') AS zone
			FROM users u
			LEFT JOIN worker_profiles wp ON wp.worker_id = u.id
			LEFT JOIN zones z ON z.id = wp.zone_id
			WHERE u.id = ?
			LIMIT 1
		`, userID).Scan(&ctx).Error; err != nil {
			return err
		}
		if ctx.ID == 0 {
			return fmt.Errorf("user not found")
		}

		var latest models.Policy
		err := tx.Where("worker_id = ?", userID).Order("id DESC").First(&latest).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			latest = models.Policy{
				WorkerID:      userID,
				Status:        "active",
				PremiumAmount: 22,
			}
			if err := tx.Create(&latest).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.Policy{}).
				Where("worker_id = ? AND id <> ? AND status = ?", userID, latest.ID, "active").
				Updates(map[string]interface{}{"status": "inactive", "updated_at": time.Now().UTC()}).
				Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Policy{}).
				Where("id = ?", latest.ID).
				Updates(map[string]interface{}{"status": "active", "updated_at": time.Now().UTC()}).
				Error; err != nil {
				return err
			}
		}

		return tx.Exec(`
			INSERT INTO active_policies (user_id, policy_id, zone, started_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id) DO UPDATE SET
				policy_id = EXCLUDED.policy_id,
				zone = EXCLUDED.zone,
				updated_at = CURRENT_TIMESTAMP
		`, userID, latest.ID, ctx.Zone).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetUserPlanStatus(userID)
}

func (s *InsurerService) EndUserPlan(userID uint) (*UserPlanStatus, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Table("users").Where("id = ?", userID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("user not found")
		}

		now := time.Now().UTC()
		if err := tx.Model(&models.Policy{}).
			Where("worker_id = ? AND status = ?", userID, "active").
			Updates(map[string]interface{}{"status": "inactive", "updated_at": now}).
			Error; err != nil {
			return err
		}

		return tx.Exec("DELETE FROM active_policies WHERE user_id = ?", userID).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetUserPlanStatus(userID)
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

func (s *InsurerService) GetMoneyExchange(levelFilter, zoneFilter string) (*MoneyExchangeSummary, error) {
	if s.DB == nil {
		return &MoneyExchangeSummary{
			PremiumPool:      2200,
			TotalSubscribed:  500,
			TotalClaims:      120,
			TotalClaimAmount: 960,
			TotalPayouts:     860,
			NetPool:          1340,
			PendingPayouts:   6,
			ZoneBreakdown: []ZoneMoneyExchange{
				{ZoneID: 1, ZoneName: "Tambaram", City: "Chennai", State: "Tamil Nadu", Level: "A", SubscribedWorkers: 58, ClaimsCount: 16, PremiumsCollected: 428, ClaimsAmount: 162, PayoutsProcessed: 143, NetFlow: 285},
			},
		}, nil
	}

	level := strings.ToUpper(strings.TrimSpace(levelFilter))
	zoneLike := strings.TrimSpace(zoneFilter)

	type zoneRow struct {
		ZoneID            uint
		ZoneName          string
		City              string
		State             string
		Level             string
		SubscribedWorkers int64
		ClaimsCount       int64
		PremiumsCollected float64
		ClaimsAmount      float64
		PayoutsProcessed  float64
	}

	rows := make([]zoneRow, 0)
	err := s.DB.Raw(`
		SELECT
			z.id AS zone_id,
			z.name AS zone_name,
			z.city,
			z.state,
			COALESCE(z.level, '') AS level,
			COALESCE(sub.subscribed_workers, 0) AS subscribed_workers,
			COALESCE(clm.claims_count, 0) AS claims_count,
			COALESCE(prem.premiums_collected, 0) AS premiums_collected,
			COALESCE(clm.claims_amount, 0) AS claims_amount,
			COALESCE(pay.payouts_processed, 0) AS payouts_processed
		FROM zones z
		LEFT JOIN (
			SELECT wp.zone_id, COUNT(DISTINCT p.worker_id) AS subscribed_workers
			FROM policies p
			JOIN worker_profiles wp ON wp.worker_id = p.worker_id
			WHERE p.status = 'active'
			GROUP BY wp.zone_id
		) sub ON sub.zone_id = z.id
		LEFT JOIN (
			SELECT wp.zone_id, COALESCE(SUM(pp.amount), 0) AS premiums_collected
			FROM premium_payments pp
			JOIN worker_profiles wp ON wp.worker_id = pp.worker_id
			WHERE pp.status IN ('completed', 'captured', 'processed')
			GROUP BY wp.zone_id
		) prem ON prem.zone_id = z.id
		LEFT JOIN (
			SELECT d.zone_id,
			       COUNT(c.id) AS claims_count,
			       COALESCE(SUM(c.claim_amount), 0) AS claims_amount
			FROM claims c
			JOIN disruptions d ON d.id = c.disruption_id
			GROUP BY d.zone_id
		) clm ON clm.zone_id = z.id
		LEFT JOIN (
			SELECT d.zone_id, COALESCE(SUM(p.amount), 0) AS payouts_processed
			FROM payouts p
			JOIN claims c ON c.id = p.claim_id
			JOIN disruptions d ON d.id = c.disruption_id
			WHERE p.status IN ('processed', 'credited', 'completed')
			GROUP BY d.zone_id
		) pay ON pay.zone_id = z.id
		WHERE (? = '' OR UPPER(COALESCE(z.level, '')) = ?)
		  AND (? = '' OR z.name ILIKE ? OR z.city ILIKE ?)
		ORDER BY z.city ASC, z.name ASC
	`, level, level, zoneLike, "%"+zoneLike+"%", "%"+zoneLike+"%").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	zoneBreakdown := make([]ZoneMoneyExchange, 0, len(rows))
	var totalSubscribed int64
	var totalClaims int64
	var premiumPool float64
	var totalClaimAmount float64
	var totalPayouts float64

	for _, row := range rows {
		net := row.PremiumsCollected - row.PayoutsProcessed
		zoneBreakdown = append(zoneBreakdown, ZoneMoneyExchange{
			ZoneID:            row.ZoneID,
			ZoneName:          row.ZoneName,
			City:              row.City,
			State:             row.State,
			Level:             row.Level,
			SubscribedWorkers: row.SubscribedWorkers,
			ClaimsCount:       row.ClaimsCount,
			PremiumsCollected: row.PremiumsCollected,
			ClaimsAmount:      row.ClaimsAmount,
			PayoutsProcessed:  row.PayoutsProcessed,
			NetFlow:           net,
		})

		totalSubscribed += row.SubscribedWorkers
		totalClaims += row.ClaimsCount
		premiumPool += row.PremiumsCollected
		totalClaimAmount += row.ClaimsAmount
		totalPayouts += row.PayoutsProcessed
	}

	var pendingPayouts int64
	pendingQuery := s.DB.Table("payouts p").
		Joins("JOIN claims c ON c.id = p.claim_id").
		Joins("JOIN disruptions d ON d.id = c.disruption_id").
		Joins("JOIN zones z ON z.id = d.zone_id").
		Where("p.status IN ?", []string{"queued", "pending", "retry_pending"})
	if level != "" {
		pendingQuery = pendingQuery.Where("UPPER(COALESCE(z.level, '')) = ?", level)
	}
	if zoneLike != "" {
		pendingQuery = pendingQuery.Where("z.name ILIKE ? OR z.city ILIKE ?", "%"+zoneLike+"%", "%"+zoneLike+"%")
	}
	_ = pendingQuery.Count(&pendingPayouts).Error

	return &MoneyExchangeSummary{
		PremiumPool:      premiumPool,
		TotalSubscribed:  totalSubscribed,
		TotalClaims:      totalClaims,
		TotalClaimAmount: totalClaimAmount,
		TotalPayouts:     totalPayouts,
		NetPool:          premiumPool - totalPayouts,
		PendingPayouts:   pendingPayouts,
		ZoneBreakdown:    zoneBreakdown,
	}, nil
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
		Take(&row).Error

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

	res := s.DB.Exec("UPDATE claims SET status = ?, fraud_verdict = ?, updated_at = ? WHERE id = ?", req.Status, req.FraudVerdict, time.Now(), claimID)
	if res.Error != nil {
		return fmt.Errorf("failed to update claim: %w", res.Error)
	}

	cid := 0
	if _, err := fmt.Sscanf(claimID, "%d", &cid); err == nil {
		audit := models.ClaimAuditLog{
			ClaimID:   uint(cid),
			Action:    "review",
			Notes:     req.Notes,
			Reviewer:  "system_user",
			CreatedAt: time.Now(),
		}

		if err := s.DB.Create(&audit).Error; err != nil {
			// Some demo databases may not include claim_audit_logs yet.
			// Keep review successful if only audit logging is unavailable.
			if !strings.Contains(strings.ToLower(err.Error()), "claim_audit_logs") {
				return fmt.Errorf("failed to create audit entry: %w", err)
			}
		}
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
		Where("COALESCE(cfs.final_verdict, 'pending') IN ('flagged', 'manual_review', 'pending', 'delay') AND c.status = 'manual_review'")

	_ = s.DB.Table("claims c").
		Joins("LEFT JOIN claim_fraud_scores cfs ON cfs.claim_id = c.id").
		Where("COALESCE(cfs.final_verdict, 'pending') IN ('flagged', 'manual_review', 'pending', 'delay') AND c.status = 'manual_review'").
		Count(&total).Error

	_ = baseQuery.Order("cfs.isolation_forest_score DESC, c.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	results := make([]models.FraudQueueItem, 0, len(rows))
	for _, row := range rows {
		var factors []models.FraudFactor
		if row.Violations != "[]" && row.Violations != "" {
			_ = json.Unmarshal([]byte(row.Violations), &factors)
		}
		var finalTags []string
		for _, f := range factors {
			finalTags = append(finalTags, strings.ToUpper(f.Name))
		}
		if len(finalTags) == 0 {
			finalTags = append(finalTags, "MANUAL_REVIEW")
		}

		results = append(results, models.FraudQueueItem{
			ClaimID:      row.ClaimID,
			Status:       "manual_review", // Contextual
			FraudVerdict: row.FinalVerdict,
			FraudScore:   row.Score,
			Violations:   finalTags,
			CreatedAt:    row.CreatedAt,
		})
	}

	return results, total, nil
}

func (s *InsurerService) GetMaintenanceChecks(offset, limit int) ([]MaintenanceCheckItem, int64, error) {
	if s.DB == nil {
		now := time.Now().UTC()
		return []MaintenanceCheckItem{{
			ID:           1,
			ClaimID:      1,
			WorkerID:     1,
			ZoneName:     "Tambaram",
			City:         "Chennai",
			Status:       "manual_review",
			FraudVerdict: "pending",
			ClaimAmount:  696,
			InitiatedAt:  now.Add(-2 * time.Hour),
			Findings:     "Awaiting reviewer response.",
		}}, 1, nil
	}

	type row struct {
		ID           uint    `gorm:"column:id"`
		ClaimID      uint    `gorm:"column:claim_id"`
		WorkerID     uint    `gorm:"column:worker_id"`
		ZoneName     string  `gorm:"column:zone_name"`
		City         string  `gorm:"column:city"`
		Status       string  `gorm:"column:status"`
		FraudVerdict string  `gorm:"column:fraud_verdict"`
		ClaimAmount  float64 `gorm:"column:claim_amount"`
		InitiatedAt  string  `gorm:"column:initiated_at"`
		ResponseAt   string  `gorm:"column:response_at"`
		Findings     string  `gorm:"column:findings"`
	}

	var rows []row
	var total int64
	_ = s.DB.Table("maintenance_check mc").
		Joins("JOIN claims c ON c.id = mc.claim_id").
		Count(&total).Error

	err := s.DB.Table("maintenance_check mc").
		Select(`
			mc.id,
			mc.claim_id,
			c.worker_id,
			COALESCE(z.name, '') AS zone_name,
			COALESCE(z.city, '') AS city,
			COALESCE(c.status, 'pending') AS status,
			COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict,
			COALESCE(c.claim_amount, 0) AS claim_amount,
			CAST(mc.initiated_date AS text) AS initiated_at,
			COALESCE(CAST(mc.response_date AS text), '') AS response_at,
			COALESCE(mc.findings, '') AS findings
		`).
		Joins("JOIN claims c ON c.id = mc.claim_id").
		Joins("LEFT JOIN disruptions d ON d.id = c.disruption_id").
		Joins("LEFT JOIN zones z ON z.id = d.zone_id").
		Order("mc.initiated_date DESC, mc.id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]MaintenanceCheckItem, 0, len(rows))
	for _, row := range rows {
		initiatedAt, _ := time.Parse(time.RFC3339Nano, row.InitiatedAt)
		if initiatedAt.IsZero() {
			initiatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", row.InitiatedAt)
		}
		var responseAt *string
		if row.ResponseAt != "" {
			resp := row.ResponseAt
			responseAt = &resp
		}
		items = append(items, MaintenanceCheckItem{
			ID:           row.ID,
			ClaimID:      row.ClaimID,
			WorkerID:     row.WorkerID,
			ZoneName:     row.ZoneName,
			City:         row.City,
			Status:       row.Status,
			FraudVerdict: row.FraudVerdict,
			ClaimAmount:  row.ClaimAmount,
			InitiatedAt:  initiatedAt,
			ResponseAt:   responseAt,
			Findings:     row.Findings,
		})
	}

	return items, total, nil
}

func (s *InsurerService) RespondToMaintenanceCheck(checkID string, findings string) error {
	if s.DB == nil {
		return nil
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		res := tx.Exec(
			"UPDATE maintenance_check SET findings = ?, response_date = ? WHERE id = ?",
			findings, now, checkID,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("maintenance check not found")
		}
		return nil
	})
}

func (s *InsurerService) GetLedger(offset, limit int) ([]models.LedgerItem, int64, error) {
	if s.DB == nil {
		return []models.LedgerItem{}, 0, nil
	}

	type rawEvent struct {
		Timestamp   time.Time
		WorkerID    uint
		Zone        string
		EventType   string
		Amount      float64
		Status      string
		ReferenceID string
	}

	var raws []rawEvent
	var total int64

	// Combined query for Premiums and Payouts
	query := `
		(SELECT 
			COALESCE(pp.payment_date, pp.created_at) AS timestamp, 
			pp.worker_id, 
			COALESCE(z.name, 'Global') AS zone, 
			'premium' AS event_type, 
			pp.amount, 
			pp.status, 
			CAST(pp.id AS TEXT) AS reference_id
		FROM premium_payments pp
		LEFT JOIN worker_profiles wp ON wp.worker_id = pp.worker_id
		LEFT JOIN zones z ON z.id = wp.zone_id
		WHERE pp.status IN ('completed', 'captured', 'processed'))
		UNION ALL
		(SELECT 
			p.created_at AS timestamp, 
			p.worker_id, 
			COALESCE(z.name, 'Global') AS zone, 
			'payout' AS event_type, 
			p.amount, 
			p.status, 
			CAST(p.claim_id AS TEXT) AS reference_id
		FROM payouts p
		LEFT JOIN claims c ON c.id = p.claim_id
		LEFT JOIN disruptions d ON d.id = c.disruption_id
		LEFT JOIN zones z ON z.id = d.zone_id
		WHERE p.status IN ('processed', 'credited', 'completed'))
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	countQuery := `
		SELECT (
			(SELECT COUNT(*) FROM premium_payments WHERE status IN ('completed', 'captured', 'processed')) + 
			(SELECT COUNT(*) FROM payouts WHERE status IN ('processed', 'credited', 'completed'))
		) AS total
	`

	err := s.DB.Raw(countQuery).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.DB.Raw(query, limit, offset).Scan(&raws).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]models.LedgerItem, 0, len(raws))
	for _, r := range raws {
		items = append(items, models.LedgerItem{
			Timestamp:   r.Timestamp,
			WorkerID:    r.WorkerID,
			Zone:        r.Zone,
			EventType:   r.EventType,
			Amount:      r.Amount,
			Status:      r.Status,
			ReferenceID: r.ReferenceID,
		})
	}

	return items, total, nil
}
