CREATE TABLE agent_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    permissions JSONB,
    metadata JSONB,
    last_accessed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_agent_access_agent_type CHECK (
        agent_type IN ('AI_AGENT', 'INSTITUTION', 'DEVELOPER', 'SERVICE')
    ),
    CONSTRAINT chk_agent_access_status CHECK (
        status IN ('ACTIVE', 'SUSPENDED', 'REVOKED', 'EXPIRED')
    )
);

CREATE INDEX idx_agent_access_user_id ON agent_access (user_id);
CREATE INDEX idx_agent_access_api_key_id ON agent_access (api_key_id);
CREATE INDEX idx_agent_access_agent_type ON agent_access (agent_type);
CREATE INDEX idx_agent_access_status ON agent_access (status);
CREATE INDEX idx_agent_access_last_accessed_at ON agent_access (last_accessed_at);
CREATE INDEX idx_agent_access_expires_at ON agent_access (expires_at);
