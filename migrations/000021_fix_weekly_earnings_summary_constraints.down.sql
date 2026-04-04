ALTER TABLE weekly_earnings_summary
DROP CONSTRAINT IF EXISTS uq_weekly_earnings_summary_worker_week_start;

-- Keep updated_at column on rollback to avoid breaking handler SQL that references it.
