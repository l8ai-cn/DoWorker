ALTER TABLE agents ADD COLUMN adapter_id VARCHAR(100);

UPDATE agents
SET adapter_id = CASE slug
  WHEN 'aider' THEN 'aider-pty'
  WHEN 'claude-code' THEN 'claude-stream-json'
  WHEN 'codex-cli' THEN 'codex-app-server'
  WHEN 'cursor-cli' THEN 'cursor-acp'
  WHEN 'do-agent' THEN 'do-agent-acp'
  WHEN 'gemini-cli' THEN 'gemini-acp'
  WHEN 'grok-build' THEN 'grok-build-acp'
  WHEN 'loopal' THEN 'loopal-acp'
  WHEN 'minimax-cli' THEN 'minimax-pty'
  WHEN 'opencode' THEN 'opencode-acp'
END
WHERE slug IN (
  'aider', 'claude-code', 'codex-cli', 'cursor-cli', 'do-agent',
  'gemini-cli', 'grok-build', 'loopal', 'minimax-cli', 'opencode'
);

ALTER TABLE agents
  ALTER COLUMN adapter_id SET NOT NULL,
  ADD CONSTRAINT agents_adapter_id_check
    CHECK (
      adapter_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
      AND char_length(adapter_id) BETWEEN 2 AND 100
    );
