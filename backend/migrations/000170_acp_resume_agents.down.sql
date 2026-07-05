UPDATE agents
SET agentfile_source = REPLACE(
        agentfile_source,
        E'CAPABILITY resume acp',
        E'CAPABILITY resume none'
    ),
    updated_at = NOW()
WHERE slug IN ('gemini-cli', 'opencode', 'loopal')
  AND is_builtin = true
  AND agentfile_source LIKE '%CAPABILITY resume acp%';
