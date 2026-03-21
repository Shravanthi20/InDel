-- migrations/000007_create_claims_payouts.up.sql
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    disruption_id INTEGER NOT NULL REFERENCES disruptions(id),
    worker_id INTEGER NOT NULL REFERENCES users(id),
    claim_amount DECIMAL(8, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    fraud_verdict VARCHAR(50),
    manual_reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_claims_worker_id_status ON claims(worker_id, status);
CREATE INDEX idx_claims_disruption_id ON claims(disruption_id);
CREATE INDEX idx_claims_fraud_verdict ON claims(fraud_verdict);

CREATE TABLE claim_fraud_scores (
    id SERIAL PRIMARY KEY,
    claim_id INTEGER NOT NULL UNIQUE REFERENCES claims(id),
    isolation_forest_score DECIMAL(5, 3),
    dbscan_score DECIMAL(5, 3),
    rule_violations JSONB,
    final_verdict VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE maintenance_check (
    id SERIAL PRIMARY KEY,
    claim_id INTEGER NOT NULL UNIQUE REFERENCES claims(id),
    initiated_date TIMESTAMP NOT NULL,
    response_date TIMESTAMP,
    findings TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payouts (
    id SERIAL PRIMARY KEY,
    claim_id INTEGER NOT NULL UNIQUE REFERENCES claims(id),
    worker_id INTEGER NOT NULL REFERENCES users(id),
    amount DECIMAL(8, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    razorpay_id VARCHAR(100),
    razorpay_status VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payouts_worker_id_status ON payouts(worker_id, status);
CREATE INDEX idx_payouts_razorpay_id ON payouts(razorpay_id);
