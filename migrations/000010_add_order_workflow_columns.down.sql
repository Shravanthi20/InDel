-- migrations/000010_add_order_workflow_columns.down.sql
DROP INDEX IF EXISTS idx_orders_updated_at;
DROP INDEX IF EXISTS idx_orders_worker_status;

ALTER TABLE orders
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS picked_up_at,
    DROP COLUMN IF EXISTS accepted_at,
    DROP COLUMN IF EXISTS distance_km,
    DROP COLUMN IF EXISTS drop_area,
    DROP COLUMN IF EXISTS pickup_area,
    DROP COLUMN IF EXISTS status;
