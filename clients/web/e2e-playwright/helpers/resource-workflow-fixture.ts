import { randomUUID } from "node:crypto";
import type { DbFixture } from "../fixtures/db.fixture";
import { TEST_ORG_SLUG, TEST_USER } from "./env";
import {
  fixtureJson,
  RESOURCE_WORKFLOW_NAME,
  RESOURCE_WORKFLOW_SLUG,
  resourceWorkflowFixtureDocuments,
} from "./resource-workflow-fixture-documents";

export interface ResourceWorkflowFixture {
  name: string;
  resourceId: string;
  slug: string;
  snapshotId: string;
  workflowId: string;
}

export function ensureResourceWorkflowFixture(
  db: DbFixture,
): ResourceWorkflowFixture {
  const existing = findFixture(db);
  if (existing) return existing;

  const coordinates = fixtureCoordinates(db);
  const resourceUID = randomUUID();
  const documents = resourceWorkflowFixtureDocuments(
    coordinates.clusterId,
    resourceUID,
  );

  const result = db.queryValue(`
    WITH snapshot AS (
      INSERT INTO worker_spec_snapshots (
        organization_id, version, spec_json, summary_json
      ) VALUES (
        ${coordinates.organizationId}, 1,
        ${fixtureJson(documents.workerSpec)},
        ${fixtureJson(documents.workerSummary)}
      )
      RETURNING id
    ), resource AS (
      INSERT INTO orchestration_resources (
        organization_id, uid, api_version, kind, namespace, name,
        display_name, labels, status, generation, resource_version,
        active_revision, created_by_id, updated_by_id
      ) VALUES (
        ${coordinates.organizationId}, '${resourceUID}',
        'agentsmesh.io/v1alpha1', 'Workflow', '${TEST_ORG_SLUG}',
        '${RESOURCE_WORKFLOW_SLUG}', '${RESOURCE_WORKFLOW_NAME}',
        '{"test-suite":"workflow-runtime"}'::jsonb, '{}'::jsonb,
        1, 1, 1, ${coordinates.actorId}, ${coordinates.actorId}
      )
      RETURNING id
    ), revision AS (
      INSERT INTO orchestration_resource_revisions (
        organization_id, resource_id, revision, generation,
        resource_version, canonical_manifest, canonical_spec,
        resolved_refs, digest, worker_spec_snapshot_id, actor_id
      )
      SELECT ${coordinates.organizationId}, resource.id, 1, 1, 1,
        ${fixtureJson(documents.canonicalManifest)},
        ${fixtureJson(documents.workflowSpec)},
        '[]'::jsonb, '${documents.digest}', snapshot.id,
        ${coordinates.actorId}
      FROM resource, snapshot
      RETURNING resource_id
    )
    INSERT INTO workflows (
      organization_id, name, slug, agent_slug, permission_mode,
      prompt_template, used_env_bundles, config_overrides,
      prompt_variables, execution_mode, autopilot_config, status,
      sandbox_strategy, session_persistence, concurrency_policy,
      max_concurrent_runs, max_retained_runs, timeout_minutes,
      idle_timeout_sec, created_by_id, worker_spec_snapshot_id,
      orchestration_resource_id, orchestration_resource_revision
    )
    SELECT ${coordinates.organizationId}, '${RESOURCE_WORKFLOW_NAME}',
      '${RESOURCE_WORKFLOW_SLUG}', 'resource-native', 'bypassPermissions',
      'echo resource workflow', ARRAY[]::text[], '{}'::jsonb,
      '{}'::jsonb, 'direct', '{}'::jsonb, 'enabled', 'fresh',
      false, 'skip', 1, 0, 1, 30, ${coordinates.actorId},
      snapshot.id, resource.id, 1
    FROM resource, snapshot, revision
    RETURNING id, orchestration_resource_id, worker_spec_snapshot_id;
  `);
  if (!result) throw new Error("failed to create resource Workflow fixture");
  return parseFixture(result);
}

export function resetResourceWorkflowFixture(db: DbFixture) {
  db.cleanup(`
    DELETE FROM workflow_runs
    WHERE workflow_id = (
      SELECT id FROM workflows
      WHERE organization_id = (
        SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      ) AND slug = '${RESOURCE_WORKFLOW_SLUG}'
    );
    UPDATE workflows
    SET total_runs = 0, successful_runs = 0, failed_runs = 0,
        last_run_at = NULL, last_pod_key = NULL, sandbox_path = NULL
    WHERE organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    ) AND slug = '${RESOURCE_WORKFLOW_SLUG}';
  `);
}

function findFixture(db: DbFixture): ResourceWorkflowFixture | null {
  const result = db.queryValue(`
    SELECT id, orchestration_resource_id, worker_spec_snapshot_id
    FROM workflows
    WHERE organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    ) AND slug = '${RESOURCE_WORKFLOW_SLUG}'
      AND orchestration_resource_revision = 1
      AND orchestration_resource_id IS NOT NULL
      AND worker_spec_snapshot_id IS NOT NULL;
  `);
  return result ? parseFixture(result) : null;
}

function fixtureCoordinates(db: DbFixture) {
  const result = db.queryValue(`
    SELECT organization.id, actor.id, runner.cluster_id
    FROM organizations organization
    JOIN users actor ON actor.username = '${TEST_USER.username}'
    JOIN LATERAL (
      SELECT cluster_id
      FROM runners
      WHERE organization_id = organization.id
        AND status IN ('online', 'busy')
        AND available_agents @> '["loopal"]'::jsonb
      ORDER BY id
      LIMIT 1
    ) runner ON true
    WHERE organization.slug = '${TEST_ORG_SLUG}';
  `);
  if (!result) {
    throw new Error("resource Workflow fixture requires an online Loopal runner");
  }
  const [organizationId, actorId, clusterId] = result.split("|");
  const parsedClusterId = Number(clusterId);
  if (!organizationId || !actorId || !Number.isSafeInteger(parsedClusterId) ||
    parsedClusterId <= 0) {
    throw new Error(`invalid resource Workflow coordinates: ${result}`);
  }
  return { organizationId, actorId, clusterId: parsedClusterId };
}

function parseFixture(result: string): ResourceWorkflowFixture {
  const [workflowId, resourceId, snapshotId] = result.split("|");
  if (!workflowId || !resourceId || !snapshotId) {
    throw new Error(`invalid resource Workflow fixture result: ${result}`);
  }
  return {
    name: RESOURCE_WORKFLOW_NAME,
    resourceId,
    slug: RESOURCE_WORKFLOW_SLUG,
    snapshotId,
    workflowId,
  };
}
