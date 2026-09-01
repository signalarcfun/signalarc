CREATE TABLE markets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    outcome_yes_label TEXT NOT NULL DEFAULT 'YES',
    outcome_no_label TEXT NOT NULL DEFAULT 'NO',
    collateral_asset TEXT NOT NULL DEFAULT 'USDC',
    chain TEXT NOT NULL,
    resolution_source TEXT,
    opens_at TIMESTAMPTZ,
    closes_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    winning_outcome TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_markets_status CHECK (
        status IN (
            'DRAFT',
            'OPEN',
            'TRADING_CLOSED',
            'RESOLUTION_PENDING',
            'RESOLVED',
            'SETTLED',
            'CANCELLED',
            'DISPUTED',
            'REFUNDED'
        )
    ),
    CONSTRAINT chk_markets_winning_outcome CHECK (
        winning_outcome IS NULL OR winning_outcome IN ('YES', 'NO')
    )
);

CREATE INDEX idx_markets_creator_user_id ON markets (creator_user_id);
CREATE INDEX idx_markets_status ON markets (status);
CREATE INDEX idx_markets_closes_at ON markets (closes_at);
