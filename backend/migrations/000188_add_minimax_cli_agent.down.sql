BEGIN;

DELETE FROM organization_agent_configs WHERE agent_slug = 'minimax-cli';
DELETE FROM organization_agents WHERE agent_slug = 'minimax-cli';
DELETE FROM user_agent_configs WHERE agent_slug = 'minimax-cli';
DELETE FROM agents WHERE slug = 'minimax-cli';

COMMIT;
