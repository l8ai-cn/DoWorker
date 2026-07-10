DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM worker_spec_snapshots LIMIT 1) THEN
        RAISE EXCEPTION
            'worker_spec_snapshots must be empty before model binding migration';
    END IF;
END
$$;

ALTER TABLE worker_spec_snapshots
    DROP CONSTRAINT worker_spec_snapshots_spec_version_matches,
    DROP CONSTRAINT worker_spec_snapshots_summary_version_matches,
    DROP CONSTRAINT worker_spec_snapshots_model_resource_id_consistent;

CREATE FUNCTION worker_spec_jsonb_is_positive_int64(value JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE
        WHEN value IS NULL OR jsonb_typeof(value) <> 'number' THEN FALSE
        WHEN value::TEXT !~ '^[1-9][0-9]*$' THEN FALSE
        WHEN length(value::TEXT) < 19 THEN TRUE
        WHEN length(value::TEXT) = 19
            THEN value::TEXT <= '9223372036854775807'
        ELSE FALSE
    END
$$;

CREATE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE
        WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
        WHEN NOT (
            binding ?& ARRAY[
                'resource_id',
                'resource_revision',
                'connection_id',
                'connection_revision',
                'provider_key',
                'model_id'
            ]
        ) THEN FALSE
        WHEN binding - ARRAY[
            'resource_id',
            'resource_revision',
            'connection_id',
            'connection_revision',
            'provider_key',
            'model_id'
        ]::TEXT[] <> '{}'::JSONB THEN FALSE
        ELSE
            worker_spec_jsonb_is_positive_int64(binding->'resource_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'resource_revision')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_revision')
            AND jsonb_typeof(binding->'provider_key') = 'string'
            AND char_length(binding->>'provider_key') BETWEEN 2 AND 100
            AND binding->>'provider_key' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
            AND jsonb_typeof(binding->'model_id') = 'string'
            AND btrim(binding->>'model_id') <> ''
    END
$$;

ALTER TABLE worker_spec_snapshots
    ADD CONSTRAINT worker_spec_snapshots_spec_version_matches
        CHECK (
            COALESCE(
                jsonb_typeof(spec_json->'version') = 'number'
                AND spec_json->>'version' = version::TEXT,
                FALSE
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_summary_version_matches
        CHECK (
            COALESCE(
                jsonb_typeof(summary_json->'version') = 'number'
                AND summary_json->>'version' = version::TEXT,
                FALSE
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_spec_model_binding_valid
        CHECK (
            worker_spec_model_binding_is_valid(
                spec_json #> '{runtime,model_binding}'
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_summary_model_binding_valid
        CHECK (
            worker_spec_model_binding_is_valid(
                summary_json->'model_binding'
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_model_binding_consistent
        CHECK (
            COALESCE(
                spec_json #> '{runtime,model_binding}'
                    = summary_json->'model_binding',
                FALSE
            )
        );
