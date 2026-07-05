ALTER TABLE permission_policies
    ADD COLUMN IF NOT EXISTS session_id VARCHAR(100) REFERENCES agent_sessions(id) ON DELETE CASCADE;

ALTER TABLE permission_policies DROP CONSTRAINT IF EXISTS permission_policies_scope_check;
ALTER TABLE permission_policies ADD CONSTRAINT permission_policies_scope_check
    CHECK (scope IN ('org', 'project', 'pod', 'session'));

CREATE INDEX IF NOT EXISTS idx_permission_policies_session
    ON permission_policies(session_id) WHERE session_id IS NOT NULL;
