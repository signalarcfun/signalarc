CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome TEXT NOT NULL,
    quantity NUMERIC(36, 18) NOT NULL DEFAULT 0,
    average_entry_price NUMERIC(36, 18) NOT NULL DEFAULT 0,
    realized_pnl NUMERIC(36, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_positions_outcome CHECK (outcome IN ('YES', 'NO')),
    CONSTRAINT chk_positions_quantity CHECK (quantity >= 0),
    CONSTRAINT uq_positions_user_market_outcome UNIQUE (user_id, market_id, outcome)
);

CREATE INDEX idx_positions_user_id ON positions (user_id);
CREATE INDEX idx_positions_market_id ON positions (market_id);
