ALTER TABLE disruptions
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_disruptions_status_processed_at
    ON disruptions(status, processed_at);
