-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(100) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(10),
    delivery JSONB NOT NULL DEFAULT '{}'::jsonb,
    payment JSONB NOT NULL DEFAULT '{}'::jsonb,
    items JSONB NOT NULL DEFAULT '[]'::jsonb,
    locale VARCHAR(10),
    internal_signature VARCHAR(100),
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(10),
    sm_id INTEGER,
    date_created TIMESTAMPTZ,
    oof_shard VARCHAR(10)
);

-- +goose Down
DROP TABLE IF EXISTS orders;
