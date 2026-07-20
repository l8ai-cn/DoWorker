import { randomUUID } from "node:crypto";
import type { DbFixture } from "../fixtures/db.fixture";
import { TEST_ORG_SLUG, TEST_USER } from "./env";
import { loopRuntimeDocuments } from "./loop-runtime-documents";
import { fixtureJson } from "./resource-workflow-fixture-documents";

export interface LoopRuntimeFixture {
  alias: string;
  goalLoopName: string;
  snapshotId: string;
  workerTemplateName: string;
}

export function createLoopRuntimeFixture(db: DbFixture): LoopRuntimeFixture {
  const suffix = Array.from(
    { length: 12 },
    () => String.fromCharCode(97 + Math.floor(Math.random() * 26)),
  ).join("");
  const alias = `循环测试环境 ${suffix}`;
  const coordinates = fixtureCoordinates(db);
  const resourceUID = randomUUID();
  const workerTemplateName = `loop-runtime-${suffix}`;
  const goalLoopName = workerTemplateName;
  const documents = loopRuntimeDocuments({
    alias,
    clusterId: coordinates.clusterId,
    resourceUID,
    workerTemplateName,
  });

  const snapshotResult = db.queryValue(`
    WITH snapshot AS (
      INSERT INTO worker_spec_snapshots (
        organization_id, version, spec_json, summary_json
      )
      VALUES (
        ${coordinates.organizationId}, 1,
        ${fixtureJson(documents.workerSpec)},
        ${fixtureJson(documents.workerSummary)}
      )
      RETURNING id
    ), resource AS (
      INSERT INTO orchestration_resources (
        organization_id, uid, api_version, kind, namespace, name, display_name,
        labels, status, created_by_id, updated_by_id
      )
      VALUES (
        ${coordinates.organizationId}, '${resourceUID}',
        'agentsmesh.io/v1alpha1', 'WorkerTemplate',
        '${TEST_ORG_SLUG}', '${workerTemplateName}', '${alias}', '{}', '{}',
        ${coordinates.actorId}, ${coordinates.actorId}
      )
      RETURNING id, organization_id, created_by_id
    ), revision AS (
      INSERT INTO orchestration_resource_revisions (
      organization_id, resource_id, revision, generation, resource_version,
      canonical_manifest, canonical_spec, resolved_refs, digest,
      worker_spec_snapshot_id, actor_id
      )
      SELECT
        resource.organization_id, resource.id, 1, 1, 1,
        ${fixtureJson(documents.canonicalManifest)},
        ${fixtureJson(documents.workerTemplateSpec)},
        '[]'::jsonb,
        '${documents.digest}',
        snapshot.id,
        resource.created_by_id
      FROM resource, snapshot
      RETURNING worker_spec_snapshot_id
    )
    SELECT worker_spec_snapshot_id FROM revision;
  `);
  const snapshotId = snapshotResult?.split(/\s+/)[0];
  if (!snapshotId) throw new Error("failed to create Loop runtime resource");
  return { alias, goalLoopName, snapshotId, workerTemplateName };
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
        AND available_agents @> '["e2e-echo"]'::jsonb
      ORDER BY id
      LIMIT 1
    ) runner ON true
    WHERE organization.slug = '${TEST_ORG_SLUG}';
  `);
  if (!result) {
    throw new Error("Loop runtime fixture requires an online E2E Echo runner");
  }
  const [organizationId, actorId, clusterId] = result.split("|");
  const parsedClusterId = Number(clusterId);
  if (!organizationId || !actorId || !Number.isSafeInteger(parsedClusterId) ||
    parsedClusterId <= 0) {
    throw new Error(`invalid Loop runtime coordinates: ${result}`);
  }
  return { organizationId, actorId, clusterId: parsedClusterId };
}

export function cleanupLoopRuntimeFixture(
  db: DbFixture,
  fixture: LoopRuntimeFixture,
) {
  db.cleanup(`
    DELETE FROM goal_loops WHERE worker_spec_snapshot_id = ${fixture.snapshotId};
  `);
}
