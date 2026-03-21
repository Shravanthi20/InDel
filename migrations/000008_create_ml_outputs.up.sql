-- migrations/000008_create_ml_outputs.up.sql
CREATE TABLE premium_model_outputs (
    id SERIAL PRIMARY KEY,
    worker_id INTEGER NOT NULL REFERENCES users(id),
    predicted_premium DECIMAL(8, 2) NOT NULL,
    features_json JSONB,
    shap_values_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_premium_model_outputs_worker_id ON premium_model_outputs(worker_id);

CREATE TABLE forecast_model_outputs (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL REFERENCES zones(id),
    forecast_date DATE NOT NULL,
    predicted_disruption_probability DECIMAL(5, 3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(zone_id, forecast_date)
);

CREATE INDEX idx_forecast_model_outputs_zone_id ON forecast_model_outputs(zone_id);
