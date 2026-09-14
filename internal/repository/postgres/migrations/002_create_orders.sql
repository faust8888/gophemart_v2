CREATE TABLE IF NOT EXISTS orders (
    number TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'NEW',
    accrual DOUBLE PRECISION,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_user_uploaded_idx ON orders (user_id, uploaded_at DESC);
