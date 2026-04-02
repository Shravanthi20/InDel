-- migrations/000005_create_orders_earnings.down.sql
DROP TABLE IF EXISTS earnings_baseline;
DROP TABLE IF EXISTS weekly_earnings_summary;
DROP TABLE IF EXISTS earnings_records;
-- Remove new columns if rolling back (if supported by DB)
ALTER TABLE orders DROP COLUMN IF EXISTS from_city;
ALTER TABLE orders DROP COLUMN IF EXISTS to_city;
ALTER TABLE orders DROP COLUMN IF EXISTS from_state;
ALTER TABLE orders DROP COLUMN IF EXISTS to_state;
ALTER TABLE orders DROP COLUMN IF EXISTS from_lat;
ALTER TABLE orders DROP COLUMN IF EXISTS from_lon;
ALTER TABLE orders DROP COLUMN IF EXISTS to_lat;
ALTER TABLE orders DROP COLUMN IF EXISTS to_lon;
DROP TABLE IF EXISTS orders;
