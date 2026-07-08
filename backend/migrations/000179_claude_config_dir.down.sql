UPDATE agents
SET agentfile_source = replace(
    agentfile_source,
    '
ENV CLAUDE_CONFIG_DIR = sandbox.root + "/claude-home"',
    ''
)
WHERE slug = 'claude-code'
  AND is_builtin = true;
