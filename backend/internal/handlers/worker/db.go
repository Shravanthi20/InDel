package worker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

var workerDB *gorm.DB

// SetDB registers the DB handle for worker handlers.
func SetDB(db *gorm.DB) {
	workerDB = db
}

// HasDB returns true if the workerDB is set (exported for other packages)
func HasDB() bool {
	return workerDB != nil
}

// GetDB returns the workerDB instance (exported for other packages)
func GetDB() *gorm.DB {
	return workerDB
}

func parseWorkerID(workerID string) (uint, error) {
	trimmed := strings.TrimSpace(workerID)
	trimmed = strings.TrimPrefix(trimmed, "worker-")
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid_worker_id")
	}
	return uint(parsed), nil
}

// EnsureDemoSeed inserts minimum worker demo data if DB is empty.
func EnsureDemoSeed() error {
	if !HasDB() {
		return nil
	}

	var user models.User
	err := workerDB.Where("phone = ?", "+919999999999").First(&user).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	zoneName := "Tambaram"
	if err := workerDB.Exec(
		"INSERT INTO zones (name, city, state, risk_rating) VALUES (?, ?, ?, ?) ON CONFLICT (name) DO NOTHING",
		zoneName, "Chennai", "Tamil Nadu", 0.62,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Create(&models.User{Phone: "+919999999999", Role: "worker"}).Error; err != nil {
		return err
	}
	if err := workerDB.Where("phone = ?", "+919999999999").First(&user).Error; err != nil {
		return err
	}

	var zone models.Zone
	if err := workerDB.Where("name = ?", zoneName).First(&zone).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id, aqi_zone, total_earnings_lifetime)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (worker_id) DO UPDATE SET
		 name = EXCLUDED.name, zone_id = EXCLUDED.zone_id, vehicle_type = EXCLUDED.vehicle_type, upi_id = EXCLUDED.upi_id`,
		user.ID, "Gayathri Worker", zone.ID, "bike", "gayathri@upi", "AQI-Medium", 125000,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		"INSERT INTO weekly_policy_cycles (week_start, week_end, policy_count, total_premium) VALUES (CURRENT_DATE, CURRENT_DATE + INTERVAL '6 day', 1, 22) ON CONFLICT (week_start, week_end) DO NOTHING",
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO policies (worker_id, status, premium_amount, policy_cycle_id)
		 VALUES (?, 'active', 22, (SELECT id FROM weekly_policy_cycles ORDER BY id DESC LIMIT 1))
		 ON CONFLICT (worker_id) WHERE status = 'active'
		 DO UPDATE SET premium_amount = EXCLUDED.premium_amount, policy_cycle_id = EXCLUDED.policy_cycle_id, updated_at = CURRENT_TIMESTAMP`,
		user.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO earnings_baseline (worker_id, baseline_amount)
		 VALUES (?, 4080)
		 ON CONFLICT (worker_id) DO UPDATE SET baseline_amount = EXCLUDED.baseline_amount, last_updated_at = CURRENT_TIMESTAMP`,
		user.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings, claim_eligible)
		 VALUES
		 (?, CURRENT_DATE - INTERVAL '21 day', CURRENT_DATE - INTERVAL '15 day', 3520, false),
		 (?, CURRENT_DATE - INTERVAL '14 day', CURRENT_DATE - INTERVAL '8 day', 3410, false),
		 (?, CURRENT_DATE - INTERVAL '7 day', CURRENT_DATE - INTERVAL '1 day', 3290, true),
		 (?, CURRENT_DATE, CURRENT_DATE + INTERVAL '6 day', 3120, true)`,
		user.ID, user.ID, user.ID, user.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO disruptions (zone_id, type, severity, signal_timestamp, confirmed_at)
		 VALUES (?, 'heavy_rain', 'high', CURRENT_TIMESTAMP - INTERVAL '6 day', CURRENT_TIMESTAMP - INTERVAL '6 day')`,
		zone.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO claims (disruption_id, worker_id, claim_amount, status, fraud_verdict)
		 VALUES ((SELECT id FROM disruptions ORDER BY id DESC LIMIT 1), ?, 870, 'approved', 'clear')`,
		user.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO payouts (claim_id, worker_id, amount, status, razorpay_id, razorpay_status)
		 VALUES ((SELECT id FROM claims ORDER BY id DESC LIMIT 1), ?, 696, 'processed', 'rzp_mock_001', 'processed')`,
		user.ID,
	).Error; err != nil {
		return err
	}

	if err := workerDB.Exec(
		`INSERT INTO orders (worker_id, zone_id, order_value)
		 VALUES (?, ?, 78), (?, ?, 62), (?, ?, 88)`,
		user.ID, zone.ID, user.ID, zone.ID, user.ID, zone.ID,
	).Error; err != nil {
		return err
	}

	return nil
}
