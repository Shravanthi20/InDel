DROP INDEX IF EXISTS idx_orders_tip_inr;

ALTER TABLE orders
    DROP COLUMN IF EXISTS tip_inr;
