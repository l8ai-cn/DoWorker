ALTER TABLE organizations ADD CONSTRAINT orchestration_organizations_id_slug_key UNIQUE (id, slug);
CREATE FUNCTION orchestration_identifier_valid(value TEXT) RETURNS BOOLEAN AS $$ SELECT
    value ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(value) BETWEEN 2 AND 100 AND value NOT IN ('about','admin','agents','api','app','auth','billing','blog','careers','changelog','dashboard','demo','docs','enterprise','false','forgot-password','invite','login','logout','me','mock-checkout','new','null','offline','onboarding','organizations','orgs','personal','popout',
     'privacy','register','reset-password','runners','settings','support','terms','true','undefined','verify-email','www') $$ LANGUAGE SQL IMMUTABLE;
CREATE TABLE orchestration_resources (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    uid UUID NOT NULL DEFAULT uuid_generate_v4(),
    api_version VARCHAR(64) NOT NULL,
    kind VARCHAR(100) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) NOT NULL DEFAULT '',
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    status JSONB NOT NULL DEFAULT '{}'::jsonb,
    generation BIGINT NOT NULL DEFAULT 1,
    resource_version BIGINT NOT NULL DEFAULT 1,
    active_revision BIGINT NOT NULL DEFAULT 1,
    created_by_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_by_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT orchestration_resources_uid_unique UNIQUE (uid),
    CONSTRAINT orchestration_resources_identity_unique UNIQUE (organization_id, api_version, kind, namespace, name),
    CONSTRAINT orchestration_resources_org_id_unique UNIQUE (organization_id, id),
    CONSTRAINT orchestration_resources_org_uid_identity_unique UNIQUE (organization_id, uid, api_version, kind, namespace, name),
    CONSTRAINT orchestration_resources_org_head_identity_unique UNIQUE (organization_id, id, uid, api_version, kind, namespace, name),
    CONSTRAINT orchestration_resources_namespace_fkey FOREIGN KEY (organization_id, namespace) REFERENCES organizations (id, slug) ON DELETE CASCADE,
    CONSTRAINT orchestration_resources_api_version_check CHECK (api_version = 'agentsmesh.io/v1alpha1'),
    CONSTRAINT orchestration_resources_kind_check CHECK (kind ~ '^[A-Z][A-Za-z0-9]{1,99}$'),
    CONSTRAINT orchestration_resources_namespace_check CHECK (orchestration_identifier_valid(namespace)),
    CONSTRAINT orchestration_resources_name_check CHECK (orchestration_identifier_valid(name)),
    CONSTRAINT orchestration_resources_labels_object CHECK (jsonb_typeof(labels) = 'object'),
    CONSTRAINT orchestration_resources_status_object CHECK (jsonb_typeof(status) = 'object'),
    CONSTRAINT orchestration_resources_timestamps_check CHECK (isfinite(created_at) AND isfinite(updated_at) AND updated_at >= created_at),
    CONSTRAINT orchestration_resources_positive_counters CHECK (generation > 0 AND generation <= active_revision AND resource_version >= active_revision)
);
CREATE INDEX idx_orchestration_resources_tenant_list ON orchestration_resources (organization_id, kind, namespace, name);
CREATE INDEX idx_orchestration_resources_tenant_head ON orchestration_resources (organization_id, updated_at DESC, id DESC);
CREATE TABLE orchestration_resource_revisions (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    resource_id BIGINT NOT NULL,
    revision BIGINT NOT NULL,
    generation BIGINT NOT NULL,
    resource_version BIGINT NOT NULL,
    canonical_manifest JSONB NOT NULL,
    canonical_spec JSONB NOT NULL,
    resolved_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    digest VARCHAR(71) NOT NULL,
    worker_spec_snapshot_id BIGINT,
    actor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT orchestration_resource_revisions_resource_revision_unique UNIQUE (resource_id, revision),
    CONSTRAINT orchestration_resource_revisions_result_unique UNIQUE (organization_id, resource_id, revision, resource_version),
    CONSTRAINT orchestration_resource_revisions_resource_fkey FOREIGN KEY (organization_id, resource_id) REFERENCES orchestration_resources (organization_id, id) ON DELETE CASCADE,
    CONSTRAINT orchestration_resource_revisions_snapshot_fkey FOREIGN KEY (organization_id, worker_spec_snapshot_id) REFERENCES worker_spec_snapshots (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT orchestration_resource_revisions_positive_counters CHECK (revision > 0 AND generation > 0 AND generation <= revision AND resource_version >= revision),
    CONSTRAINT orchestration_resource_revisions_manifest_object CHECK (jsonb_typeof(canonical_manifest) = 'object'),
    CONSTRAINT orchestration_resource_revisions_spec_object CHECK (jsonb_typeof(canonical_spec) = 'object'),
    CONSTRAINT orchestration_resource_revisions_refs_array CHECK (jsonb_typeof(resolved_refs) = 'array'),
    CONSTRAINT orchestration_resource_revisions_created_at_check CHECK (isfinite(created_at)),
    CONSTRAINT orchestration_resource_revisions_digest_check CHECK (digest ~ '^sha256:[0-9a-f]{64}$')
);
CREATE INDEX idx_orchestration_resource_revisions_history ON orchestration_resource_revisions (organization_id, resource_id, revision DESC);
ALTER TABLE orchestration_resources ADD CONSTRAINT orchestration_resources_active_revision_fkey FOREIGN KEY (id, active_revision) REFERENCES orchestration_resource_revisions (resource_id, revision) DEFERRABLE INITIALLY DEFERRED;
CREATE TABLE orchestration_resource_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_resource_id BIGINT,
    target_api_version VARCHAR(64) NOT NULL,
    target_kind VARCHAR(100) NOT NULL,
    target_namespace VARCHAR(100) NOT NULL,
    target_name VARCHAR(100) NOT NULL,
    operation VARCHAR(16) NOT NULL,
    base_head_uid UUID,
    base_resource_version BIGINT,
    draft_hash VARCHAR(71) NOT NULL,
    plan_hash VARCHAR(71) NOT NULL,
    canonical_manifest JSONB NOT NULL,
    resolved_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    semantic_diff JSONB NOT NULL DEFAULT '[]'::jsonb,
    issues JSONB NOT NULL DEFAULT '[]'::jsonb,
    artifact_kind VARCHAR(100) NOT NULL,
    artifact_json JSONB NOT NULL,
    artifact_digest VARCHAR(71) NOT NULL,
    options_revision VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    consumed_by_id BIGINT REFERENCES users(id) ON DELETE RESTRICT,
    consumption_result VARCHAR(16),
    result_resource_id BIGINT,
    result_resource_uid UUID,
    result_resource_version BIGINT,
    result_revision BIGINT,
    result_json JSONB,
    CONSTRAINT orchestration_resource_plans_namespace_fkey FOREIGN KEY (organization_id, target_namespace) REFERENCES organizations (id, slug) ON DELETE CASCADE,
    CONSTRAINT orchestration_resource_plans_base_head_fkey FOREIGN KEY (organization_id, target_resource_id, base_head_uid, target_api_version, target_kind, target_namespace, target_name)
        REFERENCES orchestration_resources (organization_id, id, uid, api_version, kind, namespace, name) ON DELETE CASCADE,
    CONSTRAINT orchestration_resource_plans_result_fkey FOREIGN KEY (organization_id, result_resource_id, result_resource_uid, target_api_version, target_kind, target_namespace, target_name)
        REFERENCES orchestration_resources (organization_id, id, uid, api_version, kind, namespace, name) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT orchestration_resource_plans_result_revision_fkey FOREIGN KEY (organization_id, result_resource_id, result_revision, result_resource_version)
        REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision, resource_version) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT orchestration_resource_plans_type_meta_check CHECK (target_api_version = 'agentsmesh.io/v1alpha1' AND target_kind ~ '^[A-Z][A-Za-z0-9]{1,99}$'),
    CONSTRAINT orchestration_resource_plans_identifiers_check CHECK (orchestration_identifier_valid(target_namespace) AND orchestration_identifier_valid(target_name)),
    CONSTRAINT orchestration_resource_plans_operation_check CHECK (operation IN ('create', 'update')),
    CONSTRAINT orchestration_resource_plans_base_state_check CHECK ((
        (operation = 'create' AND target_resource_id IS NULL AND base_head_uid IS NULL AND base_resource_version IS NULL)
        OR (operation = 'update' AND target_resource_id IS NOT NULL AND base_head_uid IS NOT NULL AND base_resource_version IS NOT NULL AND base_resource_version > 0)
    ) IS TRUE),
    CONSTRAINT orchestration_resource_plans_hashes_check CHECK (draft_hash ~ '^sha256:[0-9a-f]{64}$' AND plan_hash ~ '^sha256:[0-9a-f]{64}$' AND artifact_digest ~ '^sha256:[0-9a-f]{64}$'),
    CONSTRAINT orchestration_resource_plans_json_shapes_check CHECK (
        jsonb_typeof(canonical_manifest) = 'object' AND jsonb_typeof(resolved_refs) = 'array' AND jsonb_typeof(semantic_diff) = 'array'
        AND jsonb_typeof(issues) = 'array' AND jsonb_typeof(artifact_json) = 'object'),
    CONSTRAINT orchestration_resource_plans_artifact_kind_check CHECK (artifact_kind ~ '^[A-Z][A-Za-z0-9]{1,99}$'),
    CONSTRAINT orchestration_resource_plans_options_revision_check CHECK (
        char_length(options_revision) BETWEEN 1 AND 128
        AND options_revision = btrim(options_revision)
        AND options_revision !~ '[[:cntrl:]]'),
    CONSTRAINT orchestration_resource_plans_expiry_check CHECK (isfinite(created_at) AND isfinite(expires_at) AND expires_at > created_at AND (consumed_at IS NULL OR isfinite(consumed_at))),
    CONSTRAINT orchestration_resource_plans_result_enum_check CHECK
        (consumption_result IS NULL OR consumption_result IN ('applied', 'cancelled', 'expired')),
    CONSTRAINT orchestration_resource_plans_consumption_check CHECK ((
        (consumed_at IS NULL AND consumed_by_id IS NULL AND consumption_result IS NULL AND result_resource_id IS NULL AND result_resource_uid IS NULL
            AND result_resource_version IS NULL AND result_revision IS NULL AND result_json IS NULL)
        OR (consumed_at IS NOT NULL AND consumed_by_id = actor_id AND consumption_result = 'applied'
            AND result_resource_id IS NOT NULL AND result_resource_uid IS NOT NULL AND consumed_at >= created_at AND consumed_at < expires_at AND result_resource_version IS NOT NULL
            AND result_resource_version > 0 AND result_revision IS NOT NULL AND result_revision > 0 AND jsonb_typeof(result_json) = 'object')
        OR (consumed_at IS NOT NULL AND consumed_by_id = actor_id AND consumption_result IN ('cancelled', 'expired')
            AND consumed_at >= created_at AND ((consumption_result = 'cancelled' AND consumed_at < expires_at) OR (consumption_result = 'expired' AND consumed_at >= expires_at)) AND result_resource_id IS NULL AND result_resource_uid IS NULL
            AND result_resource_version IS NULL AND result_revision IS NULL AND jsonb_typeof(result_json) = 'object')
    ) IS TRUE)
);
CREATE INDEX idx_orchestration_resource_plans_expiry ON orchestration_resource_plans (organization_id, expires_at) WHERE consumed_at IS NULL;
