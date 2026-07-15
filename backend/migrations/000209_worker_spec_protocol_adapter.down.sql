DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM worker_spec_snapshots LIMIT 1) THEN
        RAISE EXCEPTION
            'worker_spec_snapshots must be empty before protocol adapter rollback';
    END IF;
END
$$;

ALTER TABLE worker_spec_snapshots
    DROP CONSTRAINT worker_spec_snapshots_tool_models_consistent,
    DROP CONSTRAINT worker_spec_snapshots_summary_tool_models_valid,
    DROP CONSTRAINT worker_spec_snapshots_spec_tool_models_valid;

DROP FUNCTION worker_spec_tool_model_bindings_are_valid(JSONB);
DROP FUNCTION worker_spec_tool_model_binding_is_valid(JSONB);

CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
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
