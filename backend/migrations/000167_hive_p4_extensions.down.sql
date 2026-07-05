ALTER TABLE permission_policies DROP COLUMN IF EXISTS max_usd;
ALTER TABLE permission_policies DROP COLUMN IF EXISTS policy_handler;
DROP TABLE IF EXISTS session_read_states;
