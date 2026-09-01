CREATE TABLE resolutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    winning_outcome TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    resolver_type TEXT,
    evidence_reference TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_resolutions_winning_outcome CHECK (
        winning_outcome IS NULL OR winning_outcome IN ('YES', 'NO')
    ),
    CONSTRAINT chk_resolutions_status CHECK (
        status IN ('PENDING', 'RESOLVED', 'DISPUTED', 'CANCELLED')
    ),
    CONSTRAINT uq_resolutions_market_id UNIQUE (market_id)
);

CREATE INDEX idx_resolutions_status ON resolutions (status);
CREATE INDEX idx_resolutions_resolved_at ON resolutions (resolved_at);
