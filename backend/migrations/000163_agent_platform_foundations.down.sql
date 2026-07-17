ALTER TABLE pods DROP COLUMN IF EXISTS external_session_id;
DROP TABLE IF EXISTS model_prices;
DROP TABLE IF EXISTS permission_policies;

UPDATE agents SET agentfile_source = regexp_replace(agentfile_source, E'\nCAPABILITY [^\n]+\n', E'\n', 'g'), updated_at = NOW()
WHERE slug IN ('claude-code', 'codex-cli', 'gemini-cli', 'opencode', 'cursor-cli', 'loopal')
  AND is_builtin = true;
