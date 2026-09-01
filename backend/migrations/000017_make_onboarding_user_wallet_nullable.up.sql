ALTER TABLE agent_onboarding_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_onboarding_sessions_user_wallet_non_empty;

ALTER TABLE agent_onboarding_sessions
    ALTER COLUMN user_wallet DROP NOT NULL;
