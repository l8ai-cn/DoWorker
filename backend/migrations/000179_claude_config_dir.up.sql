UPDATE agents
SET agentfile_source = replace(
    agentfile_source,
    'ENV ANTHROPIC_BASE_URL TEXT OPTIONAL',
    'ENV ANTHROPIC_BASE_URL TEXT OPTIONAL
ENV CLAUDE_CONFIG_DIR = sandbox.root + "/claude-home"'
)
WHERE slug = 'claude-code'
  AND is_builtin = true
  AND agentfile_source NOT LIKE '%CLAUDE_CONFIG_DIR%';
