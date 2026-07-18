BEGIN;

DELETE FROM organization_agent_configs WHERE agent_slug = 'seedance-expert';
DELETE FROM organization_agents WHERE agent_slug = 'seedance-expert';
DELETE FROM user_agent_configs WHERE agent_slug = 'seedance-expert';
DELETE FROM agents WHERE slug = 'seedance-expert';

COMMIT;
