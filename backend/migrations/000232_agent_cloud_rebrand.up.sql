-- Agent Cloud rebrand: accept/emit new identifiers while migrating stored brand values.

ALTER TABLE orchestration_resources
  DROP CONSTRAINT IF EXISTS orchestration_resources_api_version_check;
ALTER TABLE orchestration_resources
  ADD CONSTRAINT orchestration_resources_api_version_check
  CHECK (api_version IN ('agentcloud.io/v1alpha1', 'agentsmesh.io/v1alpha1'));

ALTER TABLE orchestration_resource_plans
  DROP CONSTRAINT IF EXISTS orchestration_resource_plans_type_meta_check;
ALTER TABLE orchestration_resource_plans
  ADD CONSTRAINT orchestration_resource_plans_type_meta_check
  CHECK (
    target_api_version IN ('agentcloud.io/v1alpha1', 'agentsmesh.io/v1alpha1')
    AND target_kind ~ '^[A-Z][A-Za-z0-9]{1,99}$'
  );

UPDATE orchestration_resources
SET api_version = 'agentcloud.io/v1alpha1'
WHERE api_version = 'agentsmesh.io/v1alpha1';

UPDATE orchestration_resource_plans
SET target_api_version = 'agentcloud.io/v1alpha1'
WHERE target_api_version = 'agentsmesh.io/v1alpha1';

UPDATE agents
SET agentfile_source = replace(
  replace(
    replace(agentfile_source, 'agentsmesh-plugin', 'agentcloud-plugin'),
    'name: "agentsmesh"',
    'name: "agentcloud"'
  ),
  'AgentsMesh collaboration plugin',
  'Agent Cloud collaboration plugin'
)
WHERE agentfile_source LIKE '%agentsmesh%';
