DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM worker_spec_snapshots
        WHERE
            (
                spec_json #> '{runtime,model_binding}' <> '{}'::JSONB
                AND NOT COALESCE(
                    (spec_json #> '{runtime,model_binding}') ? 'protocol_adapter',
                    FALSE
                )
            )
            OR (
                summary_json->'model_binding' <> '{}'::JSONB
                AND NOT COALESCE(
                    (summary_json->'model_binding') ? 'protocol_adapter',
                    FALSE
                )
            )
    ) THEN
        RAISE EXCEPTION
            'worker_spec_snapshots require protocol_adapter backfill before migration';
    END IF;
END
$$;

CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE
        WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
        WHEN binding = '{}'::JSONB THEN TRUE
        WHEN NOT (
            binding ?& ARRAY[
                'resource_id',
                'resource_revision',
                'connection_id',
                'connection_revision',
                'provider_key',
                'protocol_adapter',
                'model_id'
            ]
        ) THEN FALSE
        WHEN binding - ARRAY[
            'resource_id',
            'resource_revision',
            'connection_id',
            'connection_revision',
            'provider_key',
            'protocol_adapter',
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
            AND jsonb_typeof(binding->'protocol_adapter') = 'string'
            AND char_length(binding->>'protocol_adapter') BETWEEN 2 AND 100
            AND binding->>'protocol_adapter' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
            AND jsonb_typeof(binding->'model_id') = 'string'
            AND btrim(binding->>'model_id') <> ''
    END
$$;

CREATE FUNCTION worker_spec_tool_model_binding_is_valid(binding JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE
        WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
        WHEN NOT (
            binding ?& ARRAY[
                'role',
                'model_binding',
                'modality',
                'capability',
                'environment'
            ]
        ) THEN FALSE
        WHEN binding - ARRAY[
            'role',
            'model_binding',
            'modality',
            'capability',
            'environment'
        ]::TEXT[] <> '{}'::JSONB THEN FALSE
        ELSE
            jsonb_typeof(binding->'role') = 'string'
            AND char_length(binding->>'role') BETWEEN 2 AND 100
            AND binding->>'role' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
            AND (binding->'model_binding') <> '{}'::JSONB
            AND (binding->'model_binding') ? 'protocol_adapter'
            AND worker_spec_model_binding_is_valid(binding->'model_binding')
            AND binding->>'modality' IN (
                'chat', 'image', 'audio', 'video', 'embedding', 'multimodal'
            )
            AND binding->>'capability' IN (
                'text-generation',
                'vision-input',
                'image-generation',
                'speech-to-text',
                'text-to-speech',
                'video-generation',
                'embedding'
            )
            AND jsonb_typeof(binding->'environment') = 'object'
            AND (binding->'environment') ?& ARRAY['api_key', 'base_url', 'model_id']
            AND (binding->'environment') - ARRAY[
                'api_key', 'base_url', 'model_id'
            ]::TEXT[] = '{}'::JSONB
            AND binding->'environment'->>'api_key' ~ '^[A-Z][A-Z0-9_]*$'
            AND binding->'environment'->>'base_url' ~ '^[A-Z][A-Z0-9_]*$'
            AND binding->'environment'->>'model_id' ~ '^[A-Z][A-Z0-9_]*$'
    END
$$;

CREATE FUNCTION worker_spec_tool_model_bindings_are_valid(bindings JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE
        WHEN bindings IS NULL THEN TRUE
        WHEN jsonb_typeof(bindings) <> 'array' THEN FALSE
        WHEN EXISTS (
            SELECT 1
            FROM jsonb_array_elements(bindings) AS item
            WHERE NOT worker_spec_tool_model_binding_is_valid(item)
        ) THEN FALSE
        WHEN EXISTS (
            SELECT 1
            FROM jsonb_array_elements(bindings) AS item
            GROUP BY item->>'role'
            HAVING count(*) > 1
        ) THEN FALSE
        WHEN EXISTS (
            SELECT 1
            FROM jsonb_array_elements(bindings) AS item
            CROSS JOIN LATERAL (
                VALUES
                    (item->'environment'->>'api_key'),
                    (item->'environment'->>'base_url'),
                    (item->'environment'->>'model_id')
            ) AS target(value)
            GROUP BY target.value
            HAVING count(*) > 1
        ) THEN FALSE
        ELSE TRUE
    END
$$;

ALTER TABLE worker_spec_snapshots
    ADD CONSTRAINT worker_spec_snapshots_spec_tool_models_valid
        CHECK (
            worker_spec_tool_model_bindings_are_valid(
                spec_json #> '{runtime,tool_model_bindings}'
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_summary_tool_models_valid
        CHECK (
            worker_spec_tool_model_bindings_are_valid(
                summary_json->'tool_model_bindings'
            )
        ),
    ADD CONSTRAINT worker_spec_snapshots_tool_models_consistent
        CHECK (
            COALESCE(
                spec_json #> '{runtime,tool_model_bindings}',
                '[]'::JSONB
            ) = COALESCE(
                summary_json->'tool_model_bindings',
                '[]'::JSONB
            )
        );
