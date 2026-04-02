-- migrations/000005_create_orders_earnings.up.sql

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    zone_id INTEGER NOT NULL REFERENCES zones(id),
    order_value DECIMAL(8, 2) NOT NULL,
    from_city VARCHAR(100),
    to_city VARCHAR(100),
    from_state VARCHAR(100),
    to_state VARCHAR(100),
    from_lat DECIMAL(10, 6),
    from_lon DECIMAL(10, 6),
    to_lat DECIMAL(10, 6),
    to_lon DECIMAL(10, 6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_worker_id_created_at ON orders(worker_id, created_at);
CREATE INDEX idx_orders_zone_id ON orders(zone_id);

CREATE TABLE earnings_records (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    hours_worked INTEGER,
    amount_earned DECIMAL(8, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(worker_id, date)
);

CREATE INDEX idx_earnings_records_worker_id_date ON earnings_records(worker_id, date);

CREATE TABLE weekly_earnings_summary (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    total_earnings DECIMAL(12, 2) NOT NULL,
    claim_eligible BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_weekly_earnings_summary_worker_id ON weekly_earnings_summary(worker_id);

CREATE TABLE earnings_baseline (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL UNIQUE REFERENCES users(id),
    baseline_amount DECIMAL(8, 2) NOT NULL,
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
