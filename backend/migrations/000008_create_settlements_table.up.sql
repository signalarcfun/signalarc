CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_id UUID REFERENCES resolutions(id) ON DELETE SET NULL,
    outcome TEXT,
    amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'PENDING',
    tx_hash TEXT,
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_settlements_outcome CHECK (
        outcome IS NULL OR outcome IN ('YES', 'NO')
    ),
    CONSTRAINT chk_settlements_amount CHECK (amount >= 0),
    CONSTRAINT chk_settlements_status CHECK (
        status IN ('PENDING', 'SETTLED', 'FAILED', 'CANCELLED')
    )
);

CREATE INDEX idx_settlements_market_id ON settlements (market_id);
CREATE INDEX idx_settlements_user_id ON settlements (user_id);
CREATE INDEX idx_settlements_resolution_id ON settlements (resolution_id);
CREATE INDEX idx_settlements_status ON settlements (status);
CREATE INDEX idx_settlements_settled_at ON settlements (settled_at);
