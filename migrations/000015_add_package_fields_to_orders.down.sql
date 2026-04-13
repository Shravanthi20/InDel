DROP INDEX IF EXISTS idx_orders_package_size;

ALTER TABLE orders
DROP COLUMN IF EXISTS package_weight_kg,
DROP COLUMN IF EXISTS package_size;
