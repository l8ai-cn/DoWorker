DROP INDEX IF EXISTS idx_agent_sessions_project;
DROP INDEX IF EXISTS idx_agent_sessions_active_user;

ALTER TABLE agent_sessions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE agent_sessions DROP COLUMN IF EXISTS archived;
ALTER TABLE agent_sessions DROP COLUMN IF EXISTS project;
