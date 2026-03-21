-- migrations/000006_create_disruptions.up.sql
CREATE TABLE disruptions (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL REFERENCES zones(id),
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    signal_timestamp TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_disruptions_zone_id ON disruptions(zone_id);
CREATE INDEX idx_disruptions_type ON disruptions(type);

CREATE TABLE disruption_signals (
    id SERIAL PRIMARY KEY,
    disruption_id INTEGER NOT NULL REFERENCES disruptions(id),
    source VARCHAR(100),
    raw_data_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_disruption_signals_disruption_id ON disruption_signals(disruption_id);

CREATE TABLE disruption_eligibility (
    id SERIAL PRIMARY KEY,
    disruption_id INTEGER NOT NULL REFERENCES disruptions(id),
    worker_id INTEGER NOT NULL REFERENCES users(id),
    eligible_hours INTEGER,
    baseline_income DECIMAL(8, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(disruption_id, worker_id)
);

CREATE INDEX idx_disruption_eligibility_disruption_id ON disruption_eligibility(disruption_id);
CREATE INDEX idx_disruption_eligibility_worker_id ON disruption_eligibility(worker_id);
