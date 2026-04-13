DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_deleted_at;
DROP INDEX IF EXISTS idx_notifications_worker_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE notifications DROP COLUMN IF EXISTS updated_at;
