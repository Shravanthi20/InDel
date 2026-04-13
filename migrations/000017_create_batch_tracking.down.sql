DROP INDEX IF EXISTS idx_batch_orders_user_id;
DROP INDEX IF EXISTS idx_batch_orders_order_id;
DROP TABLE IF EXISTS batch_orders;

DROP INDEX IF EXISTS idx_batches_pickup_user_id;
DROP INDEX IF EXISTS idx_batches_status;
DROP TABLE IF EXISTS batches;