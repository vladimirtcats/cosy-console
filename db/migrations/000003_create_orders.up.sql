CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id),
    status      TEXT NOT NULL DEFAULT 'pending',
    note        TEXT NOT NULL DEFAULT '',
    total_cents BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
