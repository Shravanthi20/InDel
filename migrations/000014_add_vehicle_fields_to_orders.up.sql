-- Add vehicle modeling fields to orders table
ALTER TABLE orders
ADD COLUMN vehicle_type VARCHAR(32),
ADD COLUMN vehicle_capacity INTEGER,
ADD COLUMN allowed_zones VARCHAR(128);
