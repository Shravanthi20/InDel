-- migrations/000009_create_audit_tables.up.sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_worker_id ON notifications(worker_id);

CREATE TABLE fcm_tokens (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    token VARCHAR(255) NOT NULL UNIQUE,
    device_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fcm_tokens_worker_id ON fcm_tokens(worker_id);

CREATE TABLE idempotency_keys (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    request_method VARCHAR(10),
    request_path VARCHAR(255),
    result_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(key);

CREATE TABLE kafka_event_logs (
    id SERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    event_type VARCHAR(100),
    payload_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_kafka_event_logs_topic ON kafka_event_logs(topic);
CREATE INDEX idx_kafka_event_logs_created_at ON kafka_event_logs(created_at);

CREATE TABLE api_request_logs (
    id SERIAL PRIMARY KEY,
    gateway VARCHAR(50),
    endpoint VARCHAR(255),
    method VARCHAR(10),
    status_code INTEGER,
    response_time_ms INTEGER,
    user_id INTEGER REFERENCES users(id),
    request_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_request_logs_gateway ON api_request_logs(gateway);
CREATE INDEX idx_api_request_logs_user_id ON api_request_logs(user_id);
CREATE INDEX idx_api_request_logs_created_at ON api_request_logs(created_at);

CREATE TABLE auth_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id),
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);
CREATE INDEX idx_auth_tokens_expires_at ON auth_tokens(expires_at);
