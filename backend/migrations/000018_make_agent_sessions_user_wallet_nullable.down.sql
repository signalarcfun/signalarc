ALTER TABLE agent_wallets
    DROP CONSTRAINT IF EXISTS chk_agent_wallets_not_user_wallet;

UPDATE agent_wallets
SET user_wallet = ''
WHERE user_wallet IS NULL;

ALTER TABLE agent_wallets
    ALTER COLUMN user_wallet SET NOT NULL;

ALTER TABLE agent_wallets
    ADD CONSTRAINT chk_agent_wallets_not_user_wallet CHECK (
        lower(agent_wallet_address) <> lower(user_wallet)
    );

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_not_user_wallet;

UPDATE agent_sessions
SET user_wallet = ''
WHERE user_wallet IS NULL;

ALTER TABLE agent_sessions
    ALTER COLUMN user_wallet SET NOT NULL;

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_user_wallet_non_empty CHECK (
        length(btrim(user_wallet)) > 0
    );

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_not_user_wallet CHECK (
        lower(agent_wallet_address) <> lower(user_wallet)
    );
