CREATE TABLE items (
    id          UUID PRIMARY KEY,
    sku         TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
