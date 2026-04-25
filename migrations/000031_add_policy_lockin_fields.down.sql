-- migrations/000031_add_policy_lockin_fields.down.sql
DROP INDEX IF EXISTS idx_policies_idempotency_key;
DROP INDEX IF EXISTS idx_policies_lock_in_end;
ALTER TABLE policies DROP COLUMN IF EXISTS idempotency_key;
ALTER TABLE policies DROP COLUMN IF EXISTS lock_in_end_time;
ALTER TABLE policies DROP COLUMN IF EXISTS lock_in_start_time;
