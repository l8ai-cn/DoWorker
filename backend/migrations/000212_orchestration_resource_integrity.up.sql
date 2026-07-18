CREATE FUNCTION keep_orchestration_resource_identity() RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.id, NEW.organization_id, NEW.uid, NEW.api_version, NEW.kind, NEW.namespace,
        NEW.name, NEW.created_by_id, NEW.created_at) IS DISTINCT FROM
       (OLD.id, OLD.organization_id, OLD.uid, OLD.api_version, OLD.kind, OLD.namespace,
        OLD.name, OLD.created_by_id, OLD.created_at) THEN
        RAISE EXCEPTION 'orchestration resource ownership, identity, and uid are immutable';
    END IF;
    IF NEW.resource_version <> OLD.resource_version + 1 OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION 'orchestration resource version and timestamp must advance exactly once';
    END IF;
    IF NEW.active_revision = OLD.active_revision THEN
        IF NEW.generation <> OLD.generation OR (NEW.display_name, NEW.labels) IS DISTINCT FROM (OLD.display_name, OLD.labels) THEN
            RAISE EXCEPTION 'status-only updates cannot change desired resource state';
        END IF;
    ELSIF NEW.active_revision <> OLD.active_revision + 1 OR NEW.generation NOT IN (OLD.generation, OLD.generation + 1) THEN
        RAISE EXCEPTION 'orchestration resource revision and generation must advance exactly once';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE FUNCTION prevent_orchestration_resource_revision_mutation() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' AND pg_trigger_depth() > 1
        AND NOT EXISTS (SELECT 1 FROM organizations WHERE id = OLD.organization_id) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'orchestration resource revisions are immutable';
END;
$$ LANGUAGE plpgsql;
CREATE FUNCTION validate_orchestration_resource_revision_link() RETURNS TRIGGER AS $$
DECLARE linked_revision BIGINT; linked_generation BIGINT; linked_version BIGINT; previous_spec JSONB; active_spec JSONB;
BEGIN
    IF TG_TABLE_NAME = 'orchestration_resources' THEN
        SELECT generation, resource_version INTO linked_generation, linked_version FROM orchestration_resource_revisions
        WHERE resource_id = NEW.id AND revision = NEW.active_revision;
        IF NOT FOUND OR linked_generation <> NEW.generation OR linked_version > NEW.resource_version
            OR (TG_OP = 'INSERT' AND linked_version <> NEW.resource_version)
            OR (TG_OP = 'UPDATE' AND NEW.active_revision <> OLD.active_revision
                AND linked_version <> NEW.resource_version) THEN
            RAISE EXCEPTION 'orchestration resource head does not match its active revision';
        END IF;
        IF TG_OP = 'UPDATE' AND NEW.active_revision <> OLD.active_revision THEN
            SELECT previous.canonical_spec, active.canonical_spec INTO previous_spec, active_spec
            FROM orchestration_resource_revisions previous, orchestration_resource_revisions active
            WHERE previous.resource_id = NEW.id AND previous.revision = OLD.active_revision
                AND active.resource_id = NEW.id AND active.revision = NEW.active_revision;
            IF NOT FOUND OR (active_spec IS DISTINCT FROM previous_spec) <> (NEW.generation = OLD.generation + 1) THEN
                RAISE EXCEPTION 'orchestration resource generation does not match spec change';
            END IF;
        END IF;
    ELSE
        SELECT active_revision, generation, resource_version INTO linked_revision, linked_generation, linked_version
        FROM orchestration_resources WHERE id = NEW.resource_id AND organization_id = NEW.organization_id;
        IF NOT FOUND OR linked_revision <> NEW.revision OR linked_generation <> NEW.generation
            OR linked_version <> NEW.resource_version THEN
            RAISE EXCEPTION 'orchestration resource revision does not match its head';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE FUNCTION guard_orchestration_resource_plan() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.consumed_at IS NOT NULL OR NEW.consumed_by_id IS NOT NULL
            OR NEW.consumption_result IS NOT NULL OR NEW.result_resource_id IS NOT NULL OR NEW.result_resource_uid IS NOT NULL
            OR NEW.result_resource_version IS NOT NULL OR NEW.result_revision IS NOT NULL OR NEW.result_json IS NOT NULL THEN
            RAISE EXCEPTION 'orchestration resource plans must be inserted pending';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP = 'DELETE' THEN
        IF pg_trigger_depth() > 1 AND NOT EXISTS (SELECT 1 FROM organizations WHERE id = OLD.organization_id) THEN
            RETURN OLD;
        END IF;
        RAISE EXCEPTION 'orchestration resource plans cannot be deleted';
    END IF;
    IF OLD.consumed_at IS NOT NULL THEN
        RAISE EXCEPTION 'orchestration resource plans can only be consumed once';
    END IF;
    IF (NEW.id, NEW.organization_id, NEW.actor_id, NEW.target_resource_id, NEW.target_api_version,
        NEW.target_kind, NEW.target_namespace, NEW.target_name, NEW.operation, NEW.base_head_uid,
        NEW.base_resource_version, NEW.draft_hash, NEW.plan_hash, NEW.canonical_manifest,
        NEW.resolved_refs, NEW.semantic_diff, NEW.issues, NEW.artifact_kind, NEW.artifact_json,
        NEW.artifact_digest, NEW.options_revision, NEW.created_at, NEW.expires_at) IS DISTINCT FROM
       (OLD.id, OLD.organization_id, OLD.actor_id, OLD.target_resource_id, OLD.target_api_version,
        OLD.target_kind, OLD.target_namespace, OLD.target_name, OLD.operation, OLD.base_head_uid,
        OLD.base_resource_version, OLD.draft_hash, OLD.plan_hash, OLD.canonical_manifest,
        OLD.resolved_refs, OLD.semantic_diff, OLD.issues, OLD.artifact_kind, OLD.artifact_json,
        OLD.artifact_digest, OLD.options_revision, OLD.created_at, OLD.expires_at) THEN
        RAISE EXCEPTION 'orchestration resource plan payload is immutable';
    END IF;
    IF NEW.consumed_at IS NULL OR NEW.consumed_by_id IS NULL OR NEW.consumption_result IS NULL OR NEW.result_json IS NULL THEN
        RAISE EXCEPTION 'orchestration resource plan consumption must be atomic';
    END IF;
    IF NEW.consumption_result = 'applied' AND OLD.operation = 'update'
        AND NOT EXISTS (SELECT 1 FROM orchestration_resources WHERE organization_id = OLD.organization_id
            AND id = OLD.target_resource_id AND uid = OLD.base_head_uid AND resource_version = OLD.base_resource_version) THEN
        RAISE EXCEPTION 'orchestration resource plan is stale';
    END IF;
    IF NEW.consumption_result = 'applied' AND OLD.operation = 'create'
        AND EXISTS (SELECT 1 FROM orchestration_resources WHERE organization_id = OLD.organization_id
            AND api_version = OLD.target_api_version AND kind = OLD.target_kind
            AND namespace = OLD.target_namespace AND name = OLD.target_name) THEN
        RAISE EXCEPTION 'orchestration resource plan target already exists';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER orchestration_resources_keep_identity BEFORE UPDATE ON orchestration_resources FOR EACH ROW EXECUTE FUNCTION keep_orchestration_resource_identity();
CREATE TRIGGER orchestration_resource_revisions_immutable BEFORE UPDATE OR DELETE ON orchestration_resource_revisions FOR EACH ROW EXECUTE FUNCTION prevent_orchestration_resource_revision_mutation();
CREATE CONSTRAINT TRIGGER orchestration_resources_validate_active_revision AFTER INSERT OR UPDATE ON orchestration_resources DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION validate_orchestration_resource_revision_link();
CREATE CONSTRAINT TRIGGER orchestration_resource_revisions_validate_head AFTER INSERT ON orchestration_resource_revisions DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION validate_orchestration_resource_revision_link();
CREATE TRIGGER orchestration_resource_plans_guard BEFORE INSERT OR UPDATE OR DELETE ON orchestration_resource_plans FOR EACH ROW EXECUTE FUNCTION guard_orchestration_resource_plan();
