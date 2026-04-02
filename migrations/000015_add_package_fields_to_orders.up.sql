ALTER TABLE orders
ADD COLUMN IF NOT EXISTS package_size VARCHAR(32),
ADD COLUMN IF NOT EXISTS package_weight_kg DECIMAL(8,2);

CREATE INDEX IF NOT EXISTS idx_orders_package_size ON orders(package_size);
