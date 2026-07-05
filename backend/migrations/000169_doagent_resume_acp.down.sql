UPDATE agents
SET agentfile_source = regexp_replace(
        agentfile_source,
        E'CAPABILITY resume acp',
        E'CAPABILITY resume none',
        'g'
    ),
    updated_at = NOW()
WHERE slug = 'do-agent' AND is_builtin = true
  AND agentfile_source LIKE '%CAPABILITY resume acp%';
