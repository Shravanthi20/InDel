CREATE TABLE IF NOT EXISTS worker_payments (
    worker_id INTEGER PRIMARY KEY REFERENCES users(id),
    last_payment_timestamp TIMESTAMP NOT NULL,
    next_payment_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    coverage_status VARCHAR(20) NOT NULL DEFAULT 'Active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_worker_payments_last_payment ON worker_payments(last_payment_timestamp);
