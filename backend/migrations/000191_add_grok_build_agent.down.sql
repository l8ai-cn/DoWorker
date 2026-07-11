DELETE FROM organization_agent_configs WHERE agent_slug = 'grok-build';
DELETE FROM organization_agents WHERE agent_slug = 'grok-build';
DELETE FROM user_agent_configs WHERE agent_slug = 'grok-build';
DELETE FROM agents WHERE slug = 'grok-build';
