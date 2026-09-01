CREATE TABLE trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome TEXT NOT NULL,
    side TEXT NOT NULL,
    quantity NUMERIC(36, 18) NOT NULL,
    price NUMERIC(36, 18) NOT NULL,
    collateral_amount NUMERIC(36, 18) NOT NULL,
    fee_amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'PENDING',
    tx_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_trades_outcome CHECK (outcome IN ('YES', 'NO')),
    CONSTRAINT chk_trades_side CHECK (side IN ('BUY', 'SELL')),
    CONSTRAINT chk_trades_status CHECK (
        status IN ('PENDING', 'EXECUTED', 'FAILED', 'CANCELLED')
    ),
    CONSTRAINT chk_trades_quantity CHECK (quantity > 0),
    CONSTRAINT chk_trades_price CHECK (price >= 0 AND price <= 1),
    CONSTRAINT chk_trades_collateral_amount CHECK (collateral_amount >= 0),
    CONSTRAINT chk_trades_fee_amount CHECK (fee_amount >= 0)
);

CREATE INDEX idx_trades_user_id ON trades (user_id);
CREATE INDEX idx_trades_market_id ON trades (market_id);
CREATE INDEX idx_trades_status ON trades (status);
CREATE INDEX idx_trades_created_at ON trades (created_at);
