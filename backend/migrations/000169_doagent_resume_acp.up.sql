-- DoAgent supports vendor session/resume over ACP (S6.5).
UPDATE agents
SET agentfile_source = regexp_replace(
        agentfile_source,
        E'CAPABILITY resume none',
        E'CAPABILITY resume acp',
        'g'
    ),
    updated_at = NOW()
WHERE slug = 'do-agent' AND is_builtin = true
  AND agentfile_source LIKE '%CAPABILITY resume none%';
