CREATE TABLE session_comments (
    id VARCHAR(100) PRIMARY KEY
        CHECK (id ~ '^cmt_[a-z0-9]+$' AND char_length(id) BETWEEN 8 AND 100),
    session_id VARCHAR(100) NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    path VARCHAR(500) NOT NULL,
    start_index INT NOT NULL CHECK (start_index >= 0),
    end_index INT NOT NULL CHECK (end_index > start_index),
    body TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'addressed')),
    anchor_content TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_comments_session_path ON session_comments(session_id, path);

CREATE TABLE session_permissions (
    session_id VARCHAR(100) NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    level INT NOT NULL CHECK (level BETWEEN 1 AND 4),
    PRIMARY KEY (session_id, user_id)
);

CREATE INDEX idx_session_permissions_user ON session_permissions(user_id);
