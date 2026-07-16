UPDATE agents
SET agentfile_source = replace(
      agentfile_source,
      '/seedance-expert-home',
      '/do-agent-home'
    ),
    updated_at = NOW()
WHERE slug = 'seedance-expert'
  AND agentfile_source LIKE '%/seedance-expert-home%';
