-- migrations/000031_add_policy_lockin_fields.up.sql
-- Add lock-in window fields to the policies table.
-- lock_in_start_time: when the lock-in period began
-- lock_in_end_time:   when the lock-in period expires (policy becomes ACTIVE)
-- idempotency_key:    prevents duplicate policy creation on retries

ALTER TABLE policies ADD COLUMN IF NOT EXISTS lock_in_start_time TIMESTAMP;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS lock_in_end_time   TIMESTAMP;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS idempotency_key    VARCHAR(128);

-- Partial index for fast polling: only locked policies that need activation
CREATE INDEX IF NOT EXISTS idx_policies_lock_in_end
    ON policies(lock_in_end_time)
    WHERE status = 'locked';

-- Unique index on idempotency_key (non-null values only)
CREATE UNIQUE INDEX IF NOT EXISTS idx_policies_idempotency_key
    ON policies(idempotency_key)
    WHERE idempotency_key IS NOT NULL;
