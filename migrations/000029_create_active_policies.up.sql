-- Create table to track users with an active plan (protection policy)
CREATE TABLE IF NOT EXISTS active_policies (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    policy_id INTEGER NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    zone VARCHAR(32) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);
-- Index for fast lookup by zone
CREATE INDEX IF NOT EXISTS idx_active_policies_zone ON active_policies(zone);
-- Index for fast lookup by user
CREATE INDEX IF NOT EXISTS idx_active_policies_user_id ON active_policies(user_id);