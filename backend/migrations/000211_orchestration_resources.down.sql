DROP INDEX IF EXISTS idx_orchestration_resource_plans_expiry;
DROP INDEX IF EXISTS idx_orchestration_resource_revisions_history;
DROP INDEX IF EXISTS idx_orchestration_resources_tenant_head;
DROP INDEX IF EXISTS idx_orchestration_resources_tenant_list;

DROP TABLE IF EXISTS orchestration_resource_plans;
ALTER TABLE orchestration_resources
    DROP CONSTRAINT IF EXISTS orchestration_resources_active_revision_fkey;
DROP TABLE IF EXISTS orchestration_resource_revisions;
DROP TABLE IF EXISTS orchestration_resources;
DROP FUNCTION IF EXISTS orchestration_identifier_valid(TEXT);

ALTER TABLE organizations
    DROP CONSTRAINT IF EXISTS orchestration_organizations_id_slug_key;
