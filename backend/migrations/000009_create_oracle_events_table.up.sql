CREATE TABLE oracle_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    resolution_id UUID REFERENCES resolutions(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    outcome TEXT,
    source_name TEXT,
    source_reference TEXT,
    payload JSONB,
    observed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_oracle_events_event_type CHECK (
        event_type IN (
            'OBSERVATION',
            'RESULT_SUBMITTED',
            'RESULT_CONFIRMED',
            'RESULT_DISPUTED',
            'RESULT_CANCELLED'
        )
    ),
    CONSTRAINT chk_oracle_events_status CHECK (
        status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'FAILED')
    ),
    CONSTRAINT chk_oracle_events_outcome CHECK (
        outcome IS NULL OR outcome IN ('YES', 'NO')
    )
);

CREATE INDEX idx_oracle_events_market_id ON oracle_events (market_id);
CREATE INDEX idx_oracle_events_resolution_id ON oracle_events (resolution_id);
CREATE INDEX idx_oracle_events_event_type ON oracle_events (event_type);
CREATE INDEX idx_oracle_events_status ON oracle_events (status);
CREATE INDEX idx_oracle_events_observed_at ON oracle_events (observed_at);
