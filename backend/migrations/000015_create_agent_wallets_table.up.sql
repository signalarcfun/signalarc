CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE agent_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL UNIQUE,
    user_wallet TEXT NOT NULL,
    user_email TEXT,
    agent_wallet_address TEXT NOT NULL,
    wallet_provider TEXT NOT NULL,
    chain TEXT NOT NULL,
    status TEXT NOT NULL,
    allowed_actions TEXT[] NOT NULL,
    policy_metadata JSONB,
    source_client TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_agent_wallets_provider CHECK (
        wallet_provider IN ('circle_agent_wallet')
    ),
    CONSTRAINT chk_agent_wallets_chain CHECK (
        chain IN ('ARC-TESTNET')
    ),
    CONSTRAINT chk_agent_wallets_status CHECK (
        status IN ('active', 'disabled')
    ),
    CONSTRAINT chk_agent_wallets_allowed_actions_non_empty CHECK (
        array_length(allowed_actions, 1) IS NOT NULL
    ),
    CONSTRAINT chk_agent_wallets_not_deployer_resolver CHECK (
        lower(agent_wallet_address) <> lower('0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC')
    ),
    CONSTRAINT chk_agent_wallets_not_user_wallet CHECK (
        lower(agent_wallet_address) <> lower(user_wallet)
    )
);

CREATE INDEX idx_agent_wallets_user_wallet ON agent_wallets (user_wallet);
CREATE INDEX idx_agent_wallets_agent_wallet_address ON agent_wallets (agent_wallet_address);
CREATE INDEX idx_agent_wallets_status ON agent_wallets (status);
CREATE INDEX idx_agent_wallets_source_client ON agent_wallets (source_client);