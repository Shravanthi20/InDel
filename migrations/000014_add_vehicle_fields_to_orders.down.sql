-- Remove vehicle modeling fields from orders table
ALTER TABLE orders
DROP COLUMN IF EXISTS vehicle_type,
DROP COLUMN IF EXISTS vehicle_capacity,
DROP COLUMN IF EXISTS allowed_zones;
