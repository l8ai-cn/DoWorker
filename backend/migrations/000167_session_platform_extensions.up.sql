-- P4: session read-state persistence + cost budget policy handler

CREATE TABLE session_read_states (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(100) NOT NULL,
    last_seen BIGINT NOT NULL DEFAULT 0,
    unread BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, session_id),
    CONSTRAINT session_read_states_session_id_check
        CHECK (session_id ~ '^conv_[a-z0-9]+$' AND char_length(session_id) BETWEEN 8 AND 100)
);

CREATE INDEX idx_session_read_states_user ON session_read_states(user_id);

ALTER TABLE permission_policies
    ADD COLUMN IF NOT EXISTS policy_handler VARCHAR(50) NOT NULL DEFAULT 'acp_tool_rule'
        CHECK (policy_handler IN ('acp_tool_rule', 'session_cost_budget'));

ALTER TABLE permission_policies
    ADD COLUMN IF NOT EXISTS max_usd NUMERIC(12, 6);
