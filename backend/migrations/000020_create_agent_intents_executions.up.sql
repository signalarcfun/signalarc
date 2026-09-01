CREATE TABLE agent_intents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id TEXT NOT NULL UNIQUE,
    agent_id TEXT,
    agent_wallet_address TEXT,
    wallet_provider TEXT,
    source_client TEXT,
    client_request_id TEXT,
    action TEXT NOT NULL,
    status TEXT NOT NULL,
    requires_confirmation BOOLEAN NOT NULL DEFAULT true,
    user_wallet TEXT,
    market_id TEXT,
    market_contract_address TEXT,
    amount TEXT,
    outcome TEXT,
    resolver TEXT,
    collateral_token TEXT,
    close_timestamp TEXT,
    question TEXT,
    validation_result JSONB NOT NULL DEFAULT '{}'::jsonb,
    warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,

    CONSTRAINT chk_agent_intents_intent_id_non_empty CHECK (
        length(btrim(intent_id)) > 0
    ),
    CONSTRAINT chk_agent_intents_action_non_empty CHECK (
        length(btrim(action)) > 0
    ),
    CONSTRAINT chk_agent_intents_status CHECK (
        status IN ('preview', 'confirmed', 'executed', 'failed', 'rejected', 'cancelled')
    )
);

CREATE INDEX idx_agent_intents_agent_id_created_at ON agent_intents (agent_id, created_at DESC);
CREATE INDEX idx_agent_intents_status ON agent_intents (status);
CREATE INDEX idx_agent_intents_client_request ON agent_intents (agent_id, source_client, client_request_id)
    WHERE agent_id IS NOT NULL
      AND source_client IS NOT NULL
      AND client_request_id IS NOT NULL;

CREATE TABLE agent_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id TEXT NOT NULL REFERENCES agent_intents(intent_id) ON DELETE CASCADE,
    agent_id TEXT,
    action TEXT NOT NULL,
    status TEXT NOT NULL,
    execution_mode TEXT,
    network TEXT,
    agent_factory_address TEXT,
    market_contract_address TEXT,
    approve_transaction_hash TEXT,
    transaction_hash TEXT,
    broadcast_performed BOOLEAN NOT NULL DEFAULT false,
    readback JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,

    CONSTRAINT chk_agent_executions_action_non_empty CHECK (
        length(btrim(action)) > 0
    ),
    CONSTRAINT chk_agent_executions_status CHECK (
        status IN ('pending', 'executed', 'failed')
    )
);

CREATE INDEX idx_agent_executions_intent_id ON agent_executions (intent_id);
CREATE INDEX idx_agent_executions_agent_id_created_at ON agent_executions (agent_id, created_at DESC);
CREATE INDEX idx_agent_executions_status ON agent_executions (status);
