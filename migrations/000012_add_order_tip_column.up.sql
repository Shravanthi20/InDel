ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS tip_inr DECIMAL(8, 2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_orders_tip_inr ON orders(tip_inr);
