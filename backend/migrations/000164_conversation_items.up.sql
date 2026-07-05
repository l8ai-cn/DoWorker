CREATE TABLE conversation_items (
    id VARCHAR(100) PRIMARY KEY
        CHECK (id ~ '^item_[a-z0-9]+$' AND char_length(id) BETWEEN 8 AND 100),
    session_id VARCHAR(100) NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL,
    response_id VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    position BIGINT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_items_session_pos ON conversation_items(session_id, position DESC);
