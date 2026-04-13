DROP INDEX IF EXISTS idx_disruptions_status_processed_at;

ALTER TABLE disruptions
    DROP COLUMN IF EXISTS processed_at;
