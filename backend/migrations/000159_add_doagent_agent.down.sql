BEGIN;

DELETE FROM organization_agent_configs WHERE agent_slug = 'do-agent';
DELETE FROM organization_agents WHERE agent_slug = 'do-agent';
DELETE FROM user_agent_configs WHERE agent_slug = 'do-agent';
DELETE FROM agents WHERE slug = 'do-agent';

COMMIT;
