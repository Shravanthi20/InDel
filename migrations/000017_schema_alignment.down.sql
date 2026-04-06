-- migrations/000017_schema_alignment.down.sql

-- Drop indexes introduced in the up migration.
DROP INDEX IF EXISTS idx_premium_payments_policy_cycle_id;
DROP INDEX IF EXISTS idx_premium_payments_idempotency_key;
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

ALTER TABLE weekly_policy_cycles
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS premium_failures,
    DROP COLUMN IF EXISTS premiums_computed,
    DROP COLUMN IF EXISTS workers_evaluated,
    DROP COLUMN IF EXISTS cycle_id;

ALTER TABLE zones
    DROP COLUMN IF EXISTS level;

-- Drop tables introduced for alignment.
DROP TABLE IF EXISTS synthetic_generation_runs;
DROP TABLE IF EXISTS payout_attempts;
DROP TABLE IF EXISTS claim_audit_logs;
