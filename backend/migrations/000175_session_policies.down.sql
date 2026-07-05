DROP INDEX IF EXISTS idx_permission_policies_session;
ALTER TABLE permission_policies DROP CONSTRAINT IF EXISTS permission_policies_scope_check;
ALTER TABLE permission_policies ADD CONSTRAINT permission_policies_scope_check
    CHECK (scope IN ('org', 'project', 'pod'));
ALTER TABLE permission_policies DROP COLUMN IF EXISTS session_id;
