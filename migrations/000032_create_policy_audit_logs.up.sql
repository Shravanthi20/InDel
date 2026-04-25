-- migrations/000032_create_policy_audit_logs.up.sql
-- Audit trail for every policy state transition related to lock-in.

CREATE TABLE IF NOT EXISTS policy_audit_logs (
    id          SERIAL PRIMARY KEY,
    policy_id   INTEGER NOT NULL,
    worker_id   INTEGER NOT NULL,
    action      VARCHAR(50) NOT NULL,   -- e.g. 'policy_locked', 'policy_activated', 'claim_rejected_lockin', 'purchase_blocked_disruption'
    from_status VARCHAR(50),
    to_status   VARCHAR(50),
    reason      TEXT,
    metadata    JSONB,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_policy_audit_policy_id ON policy_audit_logs(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_audit_worker_id ON policy_audit_logs(worker_id);
CREATE INDEX IF NOT EXISTS idx_policy_audit_action    ON policy_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_policy_audit_created   ON policy_audit_logs(created_at DESC);
