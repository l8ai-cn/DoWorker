ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS project TEXT;
ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_agent_sessions_active_user
    ON agent_sessions(organization_id, user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_agent_sessions_project
    ON agent_sessions(organization_id, user_id, project)
    WHERE project IS NOT NULL AND deleted_at IS NULL;
