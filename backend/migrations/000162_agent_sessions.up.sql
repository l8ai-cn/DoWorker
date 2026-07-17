-- Session ↔ Pod mapping for web-user Omnigent compatibility.
CREATE TABLE agent_sessions (
    id VARCHAR(100) PRIMARY KEY
        CHECK (id ~ '^conv_[a-z0-9]+$' AND char_length(id) BETWEEN 8 AND 100),
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pod_key VARCHAR(100) NOT NULL REFERENCES pods(pod_key) ON DELETE CASCADE,
    agent_slug VARCHAR(50) NOT NULL,
    runner_node_id VARCHAR(100),
    title TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    parent_session_id VARCHAR(100) REFERENCES agent_sessions(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (pod_key)
);

CREATE INDEX idx_agent_sessions_org_user ON agent_sessions(organization_id, user_id);
CREATE INDEX idx_agent_sessions_parent ON agent_sessions(parent_session_id);
