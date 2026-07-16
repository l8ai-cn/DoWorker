UPDATE agents
SET agentfile_source = replace(
      agentfile_source,
      '/do-agent-home',
      '/seedance-expert-home'
    ),
    updated_at = NOW()
WHERE slug = 'seedance-expert'
  AND agentfile_source LIKE '%/do-agent-home%';
