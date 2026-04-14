package worker

import "github.com/gin-gonic/gin"

// GetEarnings returns weekly earnings
func GetEarnings(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var baseline float64 = 0
			var actual float64 = 0
			_ = workerDB.Raw("SELECT baseline_amount FROM earnings_baseline WHERE worker_id = ?", workerIDUint).Scan(&baseline).Error
			_ = workerDB.Raw("SELECT total_earnings FROM weekly_earnings_summary WHERE worker_id = ? AND week_start <= CURRENT_DATE AND week_end >= CURRENT_DATE LIMIT 1", workerIDUint).Scan(&actual).Error

			type historyRow struct {
				WeekStart     string  `gorm:"column:week_start"`
				TotalEarnings float64 `gorm:"column:total_earnings"`
			}
			rows := make([]historyRow, 0)
			_ = workerDB.Raw("SELECT week_start::text, total_earnings FROM weekly_earnings_summary WHERE worker_id = ? ORDER BY week_end DESC LIMIT 4", workerIDUint).Scan(&rows).Error

			history := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				history = append(history, gin.H{"week": row.WeekStart, "actual": int(row.TotalEarnings), "baseline": int(baseline)})
			}

			insight := "You are on track to meet your baseline."
			if actual < baseline*0.5 {
				insight = "Earnings are lower than baseline. Stay online for protection eligibility."
			} else if actual > baseline {
				insight = "Great job! You've exceeded your weekly baseline."
			}

			var paidAmount float64 = 0
			_ = workerDB.Raw(`
				SELECT COALESCE(SUM(c.claim_amount), 0)
				FROM claims c
				WHERE c.worker_id = ? AND c.status IN ('paid', 'approved', 'queued_for_payout')
			`, workerIDUint).Scan(&paidAmount).Error

			var todayAmount float64 = 0
			_ = workerDB.Raw("SELECT COALESCE(SUM(amount_earned), 0) FROM earnings_records WHERE worker_id = ? AND date = CURRENT_DATE", workerIDUint).Scan(&todayAmount).Error

			c.JSON(200, gin.H{
				"currency":           "INR",
				"this_week_actual":   int(actual),
				"this_week_baseline": int(baseline),
				"today_earnings":     int(todayAmount),
				"protected_income":   int(paidAmount),
				"insight":            insight,
				"history":            history,
			})
			return
		}
	}

	store.mu.RLock()
	earnings := store.data.Earnings
	store.mu.RUnlock()

	c.JSON(200, earnings)
}

// GetEarningsHistory returns monthly history
func GetEarningsHistory(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var baseline float64 = 0
			_ = workerDB.Raw("SELECT baseline_amount FROM earnings_baseline WHERE worker_id = ?", workerIDUint).Scan(&baseline).Error

			type row struct {
				WeekStart     string  `gorm:"column:week_start"`
				TotalEarnings float64 `gorm:"column:total_earnings"`
			}
			rows := make([]row, 0)
			_ = workerDB.Raw("SELECT week_start::text, total_earnings FROM weekly_earnings_summary WHERE worker_id = ? ORDER BY week_end DESC LIMIT 12", workerIDUint).Scan(&rows).Error

			history := make([]gin.H, 0, len(rows))
			for _, r := range rows {
				history = append(history, gin.H{"week": r.WeekStart, "actual": int(r.TotalEarnings), "baseline": int(baseline)})
			}
			c.JSON(200, gin.H{"history": history})
			return
		}
	}

	store.mu.RLock()
	history := store.data.Earnings["history"]
	store.mu.RUnlock()

	c.JSON(200, gin.H{"history": history})
}

// GetEarningsBaseline returns baseline only.
func GetEarningsBaseline(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var baseline float64 = 0
			_ = workerDB.Raw("SELECT baseline_amount FROM earnings_baseline WHERE worker_id = ?", workerIDUint).Scan(&baseline).Error
			c.JSON(200, gin.H{"baseline": int(baseline), "currency": "INR"})
			return
		}
	}

	store.mu.RLock()
	baseline := store.data.Earnings["this_week_baseline"]
	store.mu.RUnlock()

	c.JSON(200, gin.H{"baseline": baseline, "currency": "INR"})
}
