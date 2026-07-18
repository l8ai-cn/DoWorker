BEGIN;

ALTER TABLE workflow_runs
  ADD COLUMN execution_manifest JSONB;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM workflow_runs
    WHERE execution_manifest IS NULL
      AND finished_at IS NULL
  ) THEN
    RAISE EXCEPTION
      'workflow_runs contain active runs without execution manifests';
  END IF;
END
$$;

ALTER TABLE workflow_runs
  ADD CONSTRAINT workflow_runs_execution_manifest_check CHECK (
    (
      execution_manifest IS NULL
      AND finished_at IS NOT NULL
    )
    OR (
      orchestration_resource_id IS NOT NULL
      AND execution_manifest IS NOT NULL
      AND jsonb_typeof(execution_manifest) = 'object'
      AND execution_manifest @> '{"version": 1}'::jsonb
      AND CASE
        WHEN jsonb_typeof(execution_manifest -> 'organization_id') = 'number'
          AND execution_manifest ->> 'organization_id' ~ '^[0-9]+$'
        THEN (execution_manifest ->> 'organization_id')::NUMERIC
          BETWEEN 1 AND 9223372036854775807
          AND (execution_manifest ->> 'organization_id')::NUMERIC = organization_id
        ELSE FALSE
      END
      AND jsonb_typeof(execution_manifest -> 'workflow_name') = 'string'
      AND execution_manifest ->> 'workflow_name' <> ''
      AND jsonb_typeof(execution_manifest -> 'workflow_slug') = 'string'
      AND execution_manifest ->> 'workflow_slug' <> ''
      AND CASE
        WHEN jsonb_typeof(execution_manifest -> 'created_by_id') = 'number'
          AND execution_manifest ->> 'created_by_id' ~ '^[0-9]+$'
        THEN (execution_manifest ->> 'created_by_id')::NUMERIC
          BETWEEN 1 AND 9223372036854775807
        ELSE FALSE
      END
      AND jsonb_typeof(execution_manifest -> 'execution_mode') = 'string'
      AND execution_manifest ->> 'execution_mode' IN ('direct', 'autopilot')
      AND jsonb_typeof(execution_manifest -> 'autopilot') = 'object'
      AND jsonb_typeof(execution_manifest -> 'sandbox_strategy') = 'string'
      AND execution_manifest ->> 'sandbox_strategy' IN ('fresh', 'persistent')
      AND jsonb_typeof(execution_manifest -> 'session_persistence') = 'boolean'
      AND (
        NOT execution_manifest ? 'source_pod_key'
        OR jsonb_typeof(execution_manifest -> 'source_pod_key') = 'string'
      )
      AND (
        NOT execution_manifest ? 'callback_url'
        OR jsonb_typeof(execution_manifest -> 'callback_url') = 'string'
      )
      AND (
        NOT execution_manifest ? 'ticket_id'
        OR execution_manifest -> 'ticket_id' = 'null'::jsonb
        OR CASE
          WHEN jsonb_typeof(execution_manifest -> 'ticket_id') = 'number'
            AND execution_manifest ->> 'ticket_id' ~ '^[0-9]+$'
          THEN (execution_manifest ->> 'ticket_id')::NUMERIC
            BETWEEN 1 AND 9223372036854775807
          ELSE FALSE
        END
      )
      AND CASE
        WHEN jsonb_typeof(execution_manifest -> 'max_retained_runs') = 'number'
          AND execution_manifest ->> 'max_retained_runs' ~ '^[0-9]+$'
        THEN (execution_manifest ->> 'max_retained_runs')::NUMERIC
          BETWEEN 0 AND 2147483647
        ELSE FALSE
      END
      AND CASE
        WHEN jsonb_typeof(execution_manifest -> 'timeout_minutes') = 'number'
          AND execution_manifest ->> 'timeout_minutes' ~ '^[0-9]+$'
        THEN (execution_manifest ->> 'timeout_minutes')::NUMERIC
          BETWEEN 1 AND 2147483647
        ELSE FALSE
      END
      AND CASE
        WHEN jsonb_typeof(execution_manifest -> 'idle_timeout_seconds') = 'number'
          AND execution_manifest ->> 'idle_timeout_seconds' ~ '^[0-9]+$'
        THEN (execution_manifest ->> 'idle_timeout_seconds')::NUMERIC
          BETWEEN 0 AND 2147483647
        ELSE FALSE
      END
      AND (
        execution_manifest ->> 'sandbox_strategy' <> 'fresh'
        OR (
          execution_manifest -> 'session_persistence' = 'false'::jsonb
          AND COALESCE(execution_manifest ->> 'source_pod_key', '') = ''
        )
      )
    )
  );

COMMIT;
