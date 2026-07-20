import { randomUUID } from "node:crypto";
import type { ApiFixture } from "../fixtures/api.fixture";
import type { DbFixture } from "../fixtures/db.fixture";
import { createE2EEchoPod } from "./e2e-worker-spec";
import { TEST_ORG_SLUG, TEST_USER } from "./env";
import { unregisterE2ECreatedPod } from "./pod-cleanup";
import {
  fixtureJSON,
  RESOURCE_WORKFLOW_NAME,
  RESOURCE_WORKFLOW_SLUG,
  resourceWorkflowManifest,
} from "./resource-workflow-manifest";

export interface ResourceWorkflowFixture {
  name: string;
  resourceId: string;
  slug: string;
  snapshotId: string;
  workflowId: string;
}

export async function ensureResourceWorkflowFixture(
  db: DbFixture,
  api: ApiFixture,
): Promise<ResourceWorkflowFixture> {
  const existing = findFixture(db);
  if (existing) return existing;

  const client = await api.connect();
  const created = await createE2EEchoPod(client, {
    alias: "E2E Resource Workflow source",
    automationLevel: "autonomous",
  });
  const podKey = (created as { pod?: { podKey?: string } }).pod?.podKey;
  if (!podKey) throw new Error("resource Workflow source creation returned no pod key");

  const snapshotID = sourceSnapshot(db, podKey);
  await client.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
  unregisterE2ECreatedPod(podKey);
  return createFixture(db, snapshotID);
}

export function resetResourceWorkflowFixture(db: DbFixture): void {
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

function sourceSnapshot(db: DbFixture, podKey: string): string {
  const result = db.queryValue(`
    SELECT snapshot.id, artifact.artifact_digest
    FROM pods pod
    JOIN worker_spec_snapshots snapshot
      ON snapshot.id = pod.worker_spec_snapshot_id
    JOIN worker_spec_dependency_artifacts artifact
      ON artifact.organization_id = pod.organization_id
      AND artifact.worker_spec_snapshot_id = snapshot.id
    WHERE pod.pod_key = '${podKey}'
      AND pod.organization_id = (
        SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      );
  `);
  const [snapshotID, digest] = result?.split("|") ?? [];
  if (
    !/^[1-9][0-9]*$/.test(snapshotID ?? "") ||
    !/^sha256:[0-9a-f]{64}$/.test(digest ?? "")
  ) {
    throw new Error("resource Workflow source lacks a persisted dependency artifact");
  }
  return snapshotID;
}

function createFixture(db: DbFixture, snapshotID: string): ResourceWorkflowFixture {
  const [organizationID, actorID] = coordinates(db);
  const resourceUID = randomUUID();
  const manifest = resourceWorkflowManifest(resourceUID);
  const result = db.queryValue(`
    WITH resource AS (
      INSERT INTO orchestration_resources (
        organization_id, uid, api_version, kind, namespace, name,
        display_name, labels, status, generation, resource_version,
        active_revision, created_by_id, updated_by_id
      ) VALUES (
        ${organizationID}, '${resourceUID}',
        'agentsmesh.io/v1alpha1', 'Workflow', '${TEST_ORG_SLUG}',
        '${RESOURCE_WORKFLOW_SLUG}', '${RESOURCE_WORKFLOW_NAME}',
        '{"test-suite":"workflow-runtime"}'::jsonb, '{}'::jsonb,
        1, 1, 1, ${actorID}, ${actorID}
      )
      RETURNING id
    ), revision AS (
      INSERT INTO orchestration_resource_revisions (
        organization_id, resource_id, revision, generation,
        resource_version, canonical_manifest, canonical_spec,
        resolved_refs, digest, worker_spec_snapshot_id, actor_id
      )
      SELECT ${organizationID}, resource.id, 1, 1, 1,
        ${fixtureJSON(manifest.canonicalManifest)}, ${fixtureJSON(manifest.spec)},
        '[]'::jsonb, '${manifest.digest}', ${snapshotID}, ${actorID}
      FROM resource
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
    SELECT ${organizationID}, '${RESOURCE_WORKFLOW_NAME}',
      '${RESOURCE_WORKFLOW_SLUG}', 'resource-native', 'bypassPermissions',
      'echo resource workflow', ARRAY[]::text[], '{}'::jsonb,
      '{}'::jsonb, 'direct', '{}'::jsonb, 'enabled', 'fresh',
      false, 'skip', 1, 0, 1, 30, ${actorID},
      ${snapshotID}, resource.id, 1
    FROM resource, revision
    RETURNING id, orchestration_resource_id, worker_spec_snapshot_id;
  `);
  if (!result) throw new Error("failed to create resource Workflow fixture");
  return parseFixture(result);
}

function findFixture(db: DbFixture): ResourceWorkflowFixture | null {
  const result = db.queryValue(`
    SELECT workflow.id, workflow.orchestration_resource_id, workflow.worker_spec_snapshot_id
    FROM workflows workflow
    JOIN worker_spec_snapshots snapshot
      ON snapshot.id = workflow.worker_spec_snapshot_id
    JOIN worker_spec_dependency_artifacts artifact
      ON artifact.organization_id = workflow.organization_id
      AND artifact.worker_spec_snapshot_id = snapshot.id
    WHERE workflow.organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    ) AND workflow.slug = '${RESOURCE_WORKFLOW_SLUG}'
      AND workflow.orchestration_resource_revision = 1
      AND workflow.orchestration_resource_id IS NOT NULL
      AND snapshot.spec_json #>> '{runtime,worker_type,slug}' = 'e2e-echo';
  `);
  return result ? parseFixture(result) : null;
}

function coordinates(db: DbFixture): [string, string] {
  const result = db.queryValue(`
    SELECT organization.id, actor.id
    FROM organizations organization
    JOIN users actor ON actor.username = '${TEST_USER.username}'
    WHERE organization.slug = '${TEST_ORG_SLUG}';
  `);
  const [organizationID, actorID] = result?.split("|") ?? [];
  if (!/^[1-9][0-9]*$/.test(organizationID ?? "") || !/^[1-9][0-9]*$/.test(actorID ?? "")) {
    throw new Error(`invalid resource Workflow coordinates: ${result}`);
  }
  return [organizationID, actorID];
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
