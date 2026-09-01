CREATE TABLE agent_onboarding_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    onboarding_id TEXT NOT NULL UNIQUE,
    agent_id TEXT NOT NULL,
    user_email TEXT NOT NULL,
    user_wallet TEXT NOT NULL,
    requested_agent_wallet_address TEXT,
    source_client TEXT,
    channel TEXT,
    chain TEXT NOT NULL DEFAULT 'ARC-TESTNET',
    wallet_provider TEXT NOT NULL DEFAULT 'circle_agent_wallet',
    status TEXT NOT NULL,
    circle_request_id_hash TEXT,
    circle_request_expires_at TIMESTAMPTZ,
    failure_reason TEXT,
    policy_metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_agent_onboarding_sessions_onboarding_id_non_empty CHECK (
        length(btrim(onboarding_id)) > 0
    ),
    CONSTRAINT chk_agent_onboarding_sessions_agent_id_non_empty CHECK (
        length(btrim(agent_id)) > 0
    ),
    CONSTRAINT chk_agent_onboarding_sessions_user_email_non_empty CHECK (
        length(btrim(user_email)) > 0
    ),
    CONSTRAINT chk_agent_onboarding_sessions_user_wallet_non_empty CHECK (
        length(btrim(user_wallet)) > 0
    ),
    CONSTRAINT chk_agent_onboarding_sessions_chain CHECK (
        chain = 'ARC-TESTNET'
    ),
    CONSTRAINT chk_agent_onboarding_sessions_provider CHECK (
        wallet_provider = 'circle_agent_wallet'
    ),
    CONSTRAINT chk_agent_onboarding_sessions_status CHECK (
        status IN ('pending_otp', 'verified', 'expired', 'failed', 'cancelled')
    )
);

CREATE INDEX idx_agent_onboarding_sessions_agent_id ON agent_onboarding_sessions (agent_id);
CREATE INDEX idx_agent_onboarding_sessions_user_email ON agent_onboarding_sessions (user_email);
CREATE INDEX idx_agent_onboarding_sessions_user_wallet ON agent_onboarding_sessions (user_wallet);
CREATE INDEX idx_agent_onboarding_sessions_status ON agent_onboarding_sessions (status);
CREATE INDEX idx_agent_onboarding_sessions_source_client ON agent_onboarding_sessions (source_client);
CREATE INDEX idx_agent_onboarding_sessions_channel ON agent_onboarding_sessions (channel);
CREATE INDEX idx_agent_onboarding_sessions_created_at ON agent_onboarding_sessions (created_at);

CREATE TABLE agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id TEXT NOT NULL UNIQUE,
    agent_id TEXT NOT NULL,
    user_email TEXT NOT NULL,
    user_wallet TEXT NOT NULL,
    agent_wallet_address TEXT NOT NULL,
    wallet_provider TEXT NOT NULL DEFAULT 'circle_agent_wallet',
    chain TEXT NOT NULL DEFAULT 'ARC-TESTNET',
    status TEXT NOT NULL,
    allowed_actions TEXT[] NOT NULL,
    allowed_channels TEXT[] NOT NULL,
    session_metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_agent_sessions_session_id_non_empty CHECK (
        length(btrim(session_id)) > 0
    ),
    CONSTRAINT chk_agent_sessions_agent_id_non_empty CHECK (
        length(btrim(agent_id)) > 0
    ),
    CONSTRAINT chk_agent_sessions_user_email_non_empty CHECK (
        length(btrim(user_email)) > 0
    ),
    CONSTRAINT chk_agent_sessions_user_wallet_non_empty CHECK (
        length(btrim(user_wallet)) > 0
    ),
    CONSTRAINT chk_agent_sessions_agent_wallet_non_empty CHECK (
        length(btrim(agent_wallet_address)) > 0
    ),
    CONSTRAINT chk_agent_sessions_chain CHECK (
        chain = 'ARC-TESTNET'
    ),
    CONSTRAINT chk_agent_sessions_provider CHECK (
        wallet_provider = 'circle_agent_wallet'
    ),
    CONSTRAINT chk_agent_sessions_status CHECK (
        status IN ('active', 'disabled', 'revoked', 'expired')
    ),
    CONSTRAINT chk_agent_sessions_allowed_actions_non_empty CHECK (
        array_length(allowed_actions, 1) IS NOT NULL
    ),
    CONSTRAINT chk_agent_sessions_allowed_channels_non_empty CHECK (
        array_length(allowed_channels, 1) IS NOT NULL
    ),
    CONSTRAINT chk_agent_sessions_not_user_wallet CHECK (
        lower(agent_wallet_address) <> lower(user_wallet)
    )
);

CREATE INDEX idx_agent_sessions_agent_id ON agent_sessions (agent_id);
CREATE INDEX idx_agent_sessions_user_email ON agent_sessions (user_email);
CREATE INDEX idx_agent_sessions_user_wallet ON agent_sessions (user_wallet);
CREATE INDEX idx_agent_sessions_agent_wallet_address ON agent_sessions (agent_wallet_address);
CREATE INDEX idx_agent_sessions_status ON agent_sessions (status);
CREATE INDEX idx_agent_sessions_chain ON agent_sessions (chain);
CREATE INDEX idx_agent_sessions_created_at ON agent_sessions (created_at);
