CREATE TABLE worker_spec_dependency_artifacts (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    worker_spec_snapshot_id BIGINT NOT NULL
        REFERENCES worker_spec_snapshots(id) ON DELETE RESTRICT,
    artifact_json JSONB NOT NULL,
    artifact_digest TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT worker_spec_dependency_artifacts_org_positive
        CHECK (organization_id > 0),
    CONSTRAINT worker_spec_dependency_artifacts_snapshot_positive
        CHECK (worker_spec_snapshot_id > 0),
    CONSTRAINT worker_spec_dependency_artifacts_object
        CHECK (jsonb_typeof(artifact_json) = 'object'),
    CONSTRAINT worker_spec_dependency_artifacts_version_v1
        CHECK (artifact_json->>'version' = '1'),
    CONSTRAINT worker_spec_dependency_artifacts_org_matches
        CHECK (artifact_json->>'organization_id' = organization_id::TEXT),
    CONSTRAINT worker_spec_dependency_artifacts_digest
        CHECK (artifact_digest ~ '^sha256:[0-9a-f]{64}$'),
    CONSTRAINT worker_spec_dependency_artifacts_one_per_snapshot
        UNIQUE (organization_id, worker_spec_snapshot_id)
);

CREATE INDEX idx_worker_spec_dependency_artifacts_snapshot
    ON worker_spec_dependency_artifacts (worker_spec_snapshot_id);

CREATE FUNCTION prevent_worker_spec_dependency_artifact_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'worker_spec_dependency_artifacts are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER worker_spec_dependency_artifacts_immutable
    BEFORE UPDATE ON worker_spec_dependency_artifacts
    FOR EACH ROW
    EXECUTE FUNCTION prevent_worker_spec_dependency_artifact_update();
