-- migrations/000010_add_order_workflow_columns.up.sql
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'assigned',
    ADD COLUMN IF NOT EXISTS pickup_area VARCHAR(120),
    ADD COLUMN IF NOT EXISTS drop_area VARCHAR(120),
    ADD COLUMN IF NOT EXISTS distance_km DECIMAL(6, 2),
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS picked_up_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_orders_worker_status ON orders(worker_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_updated_at ON orders(updated_at);
