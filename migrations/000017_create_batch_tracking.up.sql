CREATE TABLE batches (
    batch_id VARCHAR(120) PRIMARY KEY,
    zone_level VARCHAR(10) NOT NULL,
    from_city VARCHAR(120) NOT NULL,
    to_city VARCHAR(120) NOT NULL,
    total_weight DECIMAL(10, 2) NOT NULL DEFAULT 0,
    order_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'Assigned',
    pickup_code VARCHAR(16) NOT NULL,
    delivery_code VARCHAR(16) NOT NULL,
    pickup_user_id INTEGER REFERENCES users(id),
    pickup_time TIMESTAMP,
    delivery_time TIMESTAMP,
    batch_earning_inr DECIMAL(12, 2) NOT NULL DEFAULT 0,
    earnings_posted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_batches_status ON batches(status);
CREATE INDEX idx_batches_pickup_user_id ON batches(pickup_user_id);

CREATE TABLE batch_orders (
    order_id VARCHAR(40) NOT NULL,
    batch_id VARCHAR(120) NOT NULL REFERENCES batches(batch_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(30) NOT NULL DEFAULT 'Assigned',
    pickup_time TIMESTAMP,
    delivery_time TIMESTAMP,
    delivery_address VARCHAR(255),
    contact_name VARCHAR(120),
    contact_phone VARCHAR(40),
    weight DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (batch_id, order_id)
);

CREATE INDEX idx_batch_orders_order_id ON batch_orders(order_id);
CREATE INDEX idx_batch_orders_user_id ON batch_orders(user_id);