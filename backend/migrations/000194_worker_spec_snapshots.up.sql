CREATE TABLE worker_spec_snapshots (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    version SMALLINT NOT NULL,
    spec_json JSONB NOT NULL,
    summary_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT worker_spec_snapshots_organization_positive
        CHECK (organization_id > 0),
    CONSTRAINT worker_spec_snapshots_version_v1
        CHECK (version = 1),
    CONSTRAINT worker_spec_snapshots_spec_object
        CHECK (jsonb_typeof(spec_json) = 'object'),
    CONSTRAINT worker_spec_snapshots_summary_object
        CHECK (jsonb_typeof(summary_json) = 'object'),
    CONSTRAINT worker_spec_snapshots_spec_version_matches
        CHECK (spec_json->>'version' = version::text),
    CONSTRAINT worker_spec_snapshots_summary_version_matches
        CHECK (summary_json->>'version' = version::text),
    CONSTRAINT worker_spec_snapshots_model_resource_id_consistent
        CHECK (
            CASE
                WHEN
                    spec_json #>> '{runtime,model_resource_id}' ~ '^[1-9][0-9]*$'
                    AND summary_json->>'model_resource_id' ~ '^[1-9][0-9]*$'
                    AND (
                        length(spec_json #>> '{runtime,model_resource_id}') < 19
                        OR (
                            length(spec_json #>> '{runtime,model_resource_id}') = 19
                            AND spec_json #>> '{runtime,model_resource_id}'
                                <= '9223372036854775807'
                        )
                    )
                    AND (
                        length(summary_json->>'model_resource_id') < 19
                        OR (
                            length(summary_json->>'model_resource_id') = 19
                            AND summary_json->>'model_resource_id'
                                <= '9223372036854775807'
                        )
                    )
                THEN
                    (spec_json #>> '{runtime,model_resource_id}')::BIGINT
                    = (summary_json->>'model_resource_id')::BIGINT
                ELSE FALSE
            END
        )
);

CREATE INDEX idx_worker_spec_snapshots_organization_created_at
    ON worker_spec_snapshots (organization_id, created_at DESC);

CREATE FUNCTION prevent_worker_spec_snapshot_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'worker_spec_snapshots are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER worker_spec_snapshots_immutable
    BEFORE UPDATE ON worker_spec_snapshots
    FOR EACH ROW
    EXECUTE FUNCTION prevent_worker_spec_snapshot_update();
