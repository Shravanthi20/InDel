ALTER TABLE orders
ADD COLUMN IF NOT EXISTS delivery_fee_inr DECIMAL(10,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS zone_route_path TEXT NOT NULL DEFAULT '["A"]';

CREATE INDEX IF NOT EXISTS idx_orders_delivery_fee_inr ON orders(delivery_fee_inr);
