UPDATE agent_onboarding_sessions
SET user_wallet = 'unknown / not documented'
WHERE user_wallet IS NULL OR length(btrim(user_wallet)) = 0;

ALTER TABLE agent_onboarding_sessions
    ALTER COLUMN user_wallet SET NOT NULL;

ALTER TABLE agent_onboarding_sessions
    ADD CONSTRAINT chk_agent_onboarding_sessions_user_wallet_non_empty CHECK (
        length(btrim(user_wallet)) > 0
    );
