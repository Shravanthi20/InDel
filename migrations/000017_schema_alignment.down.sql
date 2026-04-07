-- migrations/000017_schema_alignment.down.sql

-- Drop indexes introduced in the up migration.
DROP INDEX IF EXISTS idx_premium_payments_policy_cycle_id;
DROP INDEX IF EXISTS idx_premium_payments_idempotency_key;
DROP INDEX IF EXISTS idx_payouts_status_next_retry_at;
DROP INDEX IF EXISTS idx_payouts_idempotency_key;
DROP INDEX IF EXISTS idx_weekly_policy_cycles_cycle_id;
DROP INDEX IF EXISTS idx_synthetic_generation_runs_run_id;
DROP INDEX IF EXISTS idx_payout_attempts_payout_id;
DROP INDEX IF EXISTS idx_claim_audit_logs_created_at;
DROP INDEX IF EXISTS idx_claim_audit_logs_claim_id;
DROP INDEX IF EXISTS idx_zones_level_name;
DROP INDEX IF EXISTS idx_zones_level;

-- Remove columns introduced for alignment.
ALTER TABLE premium_payments
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS date,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS policy_cycle_id;

ALTER TABLE payouts
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS razorpay_status,
    DROP COLUMN IF EXISTS razorpay_id,
    DROP COLUMN IF EXISTS processed_at,
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS idempotency_key;

ALTER TABLE weekly_policy_cycles
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS premium_failures,
    DROP COLUMN IF EXISTS premiums_computed,
    DROP COLUMN IF EXISTS workers_evaluated,
    DROP COLUMN IF EXISTS cycle_id;

ALTER TABLE zones
    DROP COLUMN IF EXISTS level;

ALTER TABLE earnings_baseline
    DROP COLUMN IF EXISTS updated_at;

DROP VIEW IF EXISTS earnings_baselines;

ALTER TABLE weekly_earnings_summary
    DROP COLUMN IF EXISTS updated_at;

DROP VIEW IF EXISTS weekly_earnings_summaries;

-- Drop tables introduced for alignment.
DROP TABLE IF EXISTS synthetic_generation_runs;
DROP TABLE IF EXISTS payout_attempts;
DROP TABLE IF EXISTS claim_audit_logs;
