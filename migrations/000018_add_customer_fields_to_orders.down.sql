DROP INDEX IF EXISTS idx_orders_customer_contact_number;

ALTER TABLE orders
    DROP COLUMN IF EXISTS customer_name,
    DROP COLUMN IF EXISTS customer_id,
    DROP COLUMN IF EXISTS customer_contact_number,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS payment_method;