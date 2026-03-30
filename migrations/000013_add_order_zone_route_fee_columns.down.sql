DROP INDEX IF EXISTS idx_orders_delivery_fee_inr;

ALTER TABLE orders
DROP COLUMN IF EXISTS zone_route_path,
DROP COLUMN IF EXISTS delivery_fee_inr;
