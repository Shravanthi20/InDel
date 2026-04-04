ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS customer_name VARCHAR(120),
    ADD COLUMN IF NOT EXISTS customer_id VARCHAR(80),
    ADD COLUMN IF NOT EXISTS customer_contact_number VARCHAR(40),
    ADD COLUMN IF NOT EXISTS address VARCHAR(255),
    ADD COLUMN IF NOT EXISTS payment_method VARCHAR(30);

CREATE INDEX IF NOT EXISTS idx_orders_customer_contact_number ON orders(customer_contact_number);