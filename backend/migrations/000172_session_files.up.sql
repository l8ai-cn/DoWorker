CREATE TABLE session_files (
    id VARCHAR(100) PRIMARY KEY
        CHECK (id ~ '^file_[a-z0-9]+$' AND char_length(id) BETWEEN 8 AND 100),
    session_id VARCHAR(100) NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    bytes BIGINT NOT NULL CHECK (bytes >= 0),
    content_type VARCHAR(100) NOT NULL,
    minio_key VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_files_session ON session_files(session_id, created_at DESC);
CREATE UNIQUE INDEX idx_session_files_minio_key ON session_files(minio_key);
