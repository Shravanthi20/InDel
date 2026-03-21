-- migrations/000004_create_policies.up.sql
CREATE TABLE weekly_policy_cycles (
    id SERIAL PRIMARY KEY,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    policy_count INTEGER DEFAULT 0,
    total_premium DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(week_start, week_end)
);

CREATE TABLE policies (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    premium_amount DECIMAL(8, 2) NOT NULL,
    policy_cycle_id INTEGER REFERENCES weekly_policy_cycles(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policies_worker_id_status ON policies(worker_id, status);
CREATE INDEX idx_policies_policy_cycle_id ON policies(policy_cycle_id);

CREATE TABLE premium_payments (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    policy_id INTEGER REFERENCES policies(id),
    amount DECIMAL(8, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_premium_payments_worker_id ON premium_payments(worker_id);
