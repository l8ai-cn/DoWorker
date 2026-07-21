ALTER TABLE orchestration_resource_plans
  DROP CONSTRAINT IF EXISTS orchestration_resource_plans_type_meta_check;
ALTER TABLE orchestration_resource_plans
  ADD CONSTRAINT orchestration_resource_plans_type_meta_check
  CHECK (target_api_version = 'agentsmesh.io/v1alpha1' AND target_kind ~ '^[A-Z][A-Za-z0-9]{1,99}$');

ALTER TABLE orchestration_resources
  DROP CONSTRAINT IF EXISTS orchestration_resources_api_version_check;
ALTER TABLE orchestration_resources
  ADD CONSTRAINT orchestration_resources_api_version_check
  CHECK (api_version = 'agentsmesh.io/v1alpha1');

UPDATE orchestration_resource_plans
SET target_api_version = 'agentsmesh.io/v1alpha1'
WHERE target_api_version = 'agentcloud.io/v1alpha1';

UPDATE orchestration_resources
SET api_version = 'agentsmesh.io/v1alpha1'
WHERE api_version = 'agentcloud.io/v1alpha1';

UPDATE agents
SET agentfile_source = replace(
  replace(
    replace(agentfile_source, 'agentcloud-plugin', 'agentsmesh-plugin'),
    'name: "agentcloud"',
    'name: "agentsmesh"'
  ),
  'Agent Cloud collaboration plugin',
  'AgentsMesh collaboration plugin'
)
WHERE agentfile_source LIKE '%agentcloud%';
