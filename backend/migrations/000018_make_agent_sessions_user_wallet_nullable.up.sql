ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_not_user_wallet;

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_user_wallet_non_empty;

ALTER TABLE agent_sessions
    ALTER COLUMN user_wallet DROP NOT NULL;

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_not_user_wallet CHECK (
        user_wallet IS NULL OR lower(agent_wallet_address) <> lower(user_wallet)
    );

ALTER TABLE agent_wallets
    DROP CONSTRAINT IF EXISTS chk_agent_wallets_not_user_wallet;

ALTER TABLE agent_wallets
    ALTER COLUMN user_wallet DROP NOT NULL;

ALTER TABLE agent_wallets
    ADD CONSTRAINT chk_agent_wallets_not_user_wallet CHECK (
        user_wallet IS NULL OR lower(agent_wallet_address) <> lower(user_wallet)
    );
