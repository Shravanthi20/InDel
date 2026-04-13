-- Improve disruption -> worker notification read performance for frontend/API flows.
CREATE INDEX IF NOT EXISTS idx_disruptions_created_at_desc
ON disruptions (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_disruptions_zone_created_at_desc
ON disruptions (zone_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_worker_type_created_at_desc
ON notifications (worker_id, type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_type_created_at_desc
ON notifications (type, created_at DESC);
