-- Drop the active_policies table
DROP INDEX IF EXISTS idx_active_policies_zone;
DROP INDEX IF EXISTS idx_active_policies_user_id;
DROP TABLE IF EXISTS active_policies;