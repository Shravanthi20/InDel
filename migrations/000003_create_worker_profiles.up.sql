-- migrations/000003_create_worker_profiles.up.sql
CREATE TABLE worker_profiles (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL UNIQUE REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    zone_id INTEGER NOT NULL REFERENCES zones(id),
    vehicle_type VARCHAR(50),
    upi_id VARCHAR(100),
    aqi_zone VARCHAR(50),
    total_earnings_lifetime DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_worker_profiles_worker_id ON worker_profiles(worker_id);
CREATE INDEX idx_worker_profiles_zone_id ON worker_profiles(zone_id);

CREATE TABLE worker_zone_history (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    zone_id INTEGER NOT NULL REFERENCES zones(id),
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_worker_zone_history_worker_id ON worker_zone_history(worker_id);
CREATE INDEX idx_worker_zone_history_zone_id ON worker_zone_history(zone_id);
