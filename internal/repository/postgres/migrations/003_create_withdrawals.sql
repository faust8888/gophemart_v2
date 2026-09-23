CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number TEXT NOT NULL UNIQUE,
    amount DOUBLE PRECISION NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS withdrawals_user_processed_idx ON withdrawals (user_id, processed_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS withdrawals_order_number_uidx ON withdrawals (order_number);
