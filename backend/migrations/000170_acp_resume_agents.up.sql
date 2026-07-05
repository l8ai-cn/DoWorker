UPDATE agents
SET agentfile_source = REPLACE(
        agentfile_source,
        E'CAPABILITY resume none',
        E'CAPABILITY resume acp'
    ),
    updated_at = NOW()
WHERE slug IN ('gemini-cli', 'opencode', 'loopal')
  AND is_builtin = true
  AND agentfile_source LIKE '%CAPABILITY resume none%';
