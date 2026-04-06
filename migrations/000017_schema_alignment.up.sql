-- migrations/000017_schema_alignment.up.sql
-- Idempotent schema alignment for production drift.

-- 1) Zones: code expects `level`
ALTER TABLE zones
    ADD COLUMN IF NOT EXISTS level VARCHAR(10);

ALTER TABLE zones
    ALTER COLUMN level SET DEFAULT 'B';

UPDATE zones
SET level = CASE
    WHEN risk_rating >= 0.70 THEN 'C'
    WHEN risk_rating >= 0.45 THEN 'B'
    ELSE 'A'
END
WHERE level IS NULL OR level = '';

CREATE INDEX IF NOT EXISTS idx_zones_level ON zones(level);
CREATE INDEX IF NOT EXISTS idx_zones_level_name ON zones(level, name);

-- 2) claim_audit_logs: used by insurer review flow and synthetic cleanup
CREATE TABLE IF NOT EXISTS claim_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    claim_id INTEGER NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    notes TEXT,
    reviewer VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_claim_audit_logs_claim_id ON claim_audit_logs(claim_id);
CREATE INDEX IF NOT EXISTS idx_claim_audit_logs_created_at ON claim_audit_logs(created_at);

-- 3) payout_attempts: used by payout processing retries
CREATE TABLE IF NOT EXISTS payout_attempts (
    id BIGSERIAL PRIMARY KEY,
    payout_id INTEGER NOT NULL REFERENCES payouts(id) ON DELETE CASCADE,
    attempt_no INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payout_attempts_payout_id ON payout_attempts(payout_id);

-- 4) synthetic_generation_runs: used by synthetic generation audit trail
CREATE TABLE IF NOT EXISTS synthetic_generation_runs (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(120) NOT NULL UNIQUE,
    seed INTEGER NOT NULL DEFAULT 0,
    scenario VARCHAR(120) NOT NULL DEFAULT 'default',
    output_dir TEXT,
    workers_created INTEGER NOT NULL DEFAULT 0,
    zones_created INTEGER NOT NULL DEFAULT 0,
    disruptions_created INTEGER NOT NULL DEFAULT 0,
    claims_created INTEGER NOT NULL DEFAULT 0,
    payouts_created INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_synthetic_generation_runs_run_id ON synthetic_generation_runs(run_id);

-- 5) weekly_policy_cycles: align with runtime model fields
ALTER TABLE weekly_policy_cycles
    ADD COLUMN IF NOT EXISTS cycle_id VARCHAR(120),
    ADD COLUMN IF NOT EXISTS workers_evaluated INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS premiums_computed INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS premium_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'running',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE UNIQUE INDEX IF NOT EXISTS idx_weekly_policy_cycles_cycle_id ON weekly_policy_cycles(cycle_id);

-- Backfill cycle_id for old rows when absent.
UPDATE weekly_policy_cycles
SET cycle_id = CONCAT('cyc_', EXTRACT(ISOYEAR FROM week_start)::TEXT, '_w', LPAD(EXTRACT(WEEK FROM week_start)::TEXT, 2, '0'))
WHERE (cycle_id IS NULL OR cycle_id = '') AND week_start IS NOT NULL;

-- 6) premium_payments: align with runtime model fields
ALTER TABLE premium_payments
    ADD COLUMN IF NOT EXISTS policy_cycle_id INTEGER REFERENCES weekly_policy_cycles(id),
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(160),
    ADD COLUMN IF NOT EXISTS date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE UNIQUE INDEX IF NOT EXISTS idx_premium_payments_idempotency_key ON premium_payments(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_premium_payments_policy_cycle_id ON premium_payments(policy_cycle_id);
