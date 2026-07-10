DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM worker_spec_snapshots LIMIT 1) THEN
        RAISE EXCEPTION
            'worker_spec_snapshots must be empty before model binding rollback';
    END IF;
END
$$;

ALTER TABLE worker_spec_snapshots
    DROP CONSTRAINT worker_spec_snapshots_model_binding_consistent,
    DROP CONSTRAINT worker_spec_snapshots_summary_model_binding_valid,
    DROP CONSTRAINT worker_spec_snapshots_spec_model_binding_valid,
    DROP CONSTRAINT worker_spec_snapshots_summary_version_matches,
    DROP CONSTRAINT worker_spec_snapshots_spec_version_matches;

DROP FUNCTION worker_spec_model_binding_is_valid(JSONB);
DROP FUNCTION worker_spec_jsonb_is_positive_int64(JSONB);

ALTER TABLE worker_spec_snapshots
    ADD CONSTRAINT worker_spec_snapshots_spec_version_matches
        CHECK (spec_json->>'version' = version::TEXT),
    ADD CONSTRAINT worker_spec_snapshots_summary_version_matches
        CHECK (summary_json->>'version' = version::TEXT),
    ADD CONSTRAINT worker_spec_snapshots_model_resource_id_consistent
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
        );
