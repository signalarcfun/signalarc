CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    market_id UUID REFERENCES markets(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    source TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    target_url TEXT,
    payload JSONB,
    response_status INTEGER,
    error_message TEXT,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_webhooks_status CHECK (
        status IN ('PENDING', 'DELIVERED', 'FAILED', 'CANCELLED')
    ),
    CONSTRAINT chk_webhooks_attempts CHECK (attempts >= 0)
);

CREATE INDEX idx_webhooks_user_id ON webhooks (user_id);
CREATE INDEX idx_webhooks_market_id ON webhooks (market_id);
CREATE INDEX idx_webhooks_event_type ON webhooks (event_type);
CREATE INDEX idx_webhooks_source ON webhooks (source);
CREATE INDEX idx_webhooks_status ON webhooks (status);
CREATE INDEX idx_webhooks_last_attempt_at ON webhooks (last_attempt_at);
CREATE INDEX idx_webhooks_created_at ON webhooks (created_at);
