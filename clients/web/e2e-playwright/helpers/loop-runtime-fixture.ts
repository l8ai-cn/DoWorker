import type { DbFixture } from "../fixtures/db.fixture";
import type { ApiFixture } from "../fixtures/api.fixture";
import {
  buildE2EEchoWorkerSpec,
  type E2EWorkerSpecDraft,
} from "./e2e-worker-spec";
import { applyE2EWorkerTemplate } from "./e2e-worker-template-resource";
import { TEST_ORG_SLUG } from "./env";
import {
  registerE2ECreatedPod,
  unregisterE2ECreatedPod,
} from "./pod-cleanup";

export interface LoopRuntimeFixture {
  goalLoopName: string;
  optionLabel: string;
  snapshotId: string;
}

export async function createLoopRuntimeFixture(
  db: DbFixture,
  api: ApiFixture,
): Promise<LoopRuntimeFixture> {
  const suffix = Array.from(
    { length: 12 },
    () => String.fromCharCode(97 + Math.floor(Math.random() * 26)),
  ).join("");
  const client = await api.connect();
  const worker = await buildE2EEchoWorkerSpec(client, {
    alias: `Loop runtime ${suffix}`,
    automationLevel: "autonomous",
  });
  const created = await createSourcePod(client, worker);
  const sourcePodKey = (created as { pod?: { podKey?: string } }).pod?.podKey;
  if (!sourcePodKey) {
    throw new Error("Loop runtime source Worker creation returned no pod key");
  }
  const source = db.queryValue(`
    SELECT snapshot.id, snapshot.summary_json->>'alias', artifact.artifact_digest
    FROM pods pod
    JOIN worker_spec_snapshots snapshot
      ON snapshot.id = pod.worker_spec_snapshot_id
    JOIN worker_spec_dependency_artifacts artifact
      ON artifact.organization_id = pod.organization_id
      AND artifact.worker_spec_snapshot_id = snapshot.id
    WHERE pod.pod_key = '${sourcePodKey}'
      AND pod.organization_id = (
        SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
      );
  `);
  if (!source) {
    throw new Error(
      "Loop runtime source Worker lacks a persisted dependency artifact",
    );
  }
  const [snapshotId, alias, digest] = source.split("|");
  if (!snapshotId || !alias || !/^sha256:[0-9a-f]{64}$/.test(digest ?? "")) {
    throw new Error(`invalid Loop runtime source Worker artifact: ${source}`);
  }
  const templateName = `loop-runtime-${suffix}`;
  const templateSnapshotID = await applyE2EWorkerTemplate(
    client,
    templateName,
    worker,
  );
  assertTemplateArtifact(db, templateName, templateSnapshotID);
  await client.pod.terminatePod({
    orgSlug: TEST_ORG_SLUG,
    podKey: sourcePodKey,
  });
  unregisterE2ECreatedPod(sourcePodKey);
  return {
    goalLoopName: `goal-loop-${suffix}`,
    optionLabel: `${worker.alias} · WorkerTemplate · 模板 ${templateName}`,
    snapshotId: templateSnapshotID,
  };
}

export async function cleanupLoopRuntimeFixture(
  db: DbFixture,
  fixture: LoopRuntimeFixture,
): Promise<void> {
  db.cleanup(`
    DELETE FROM goal_loops WHERE worker_spec_snapshot_id = ${fixture.snapshotId};
  `);
}

function createSourcePod(
  client: Awaited<ReturnType<ApiFixture["connect"]>>,
  worker: E2EWorkerSpecDraft,
) {
  return client.pod.createPod({
    orgSlug: TEST_ORG_SLUG,
    cols: 80,
    rows: 24,
    workerSpec: worker,
  }).then((created) => {
    const podKey = (created as { pod?: { podKey?: string } }).pod?.podKey;
    if (!podKey) {
      throw new Error("CreatePod returned no pod key for E2E cleanup registration");
    }
    registerE2ECreatedPod(podKey, worker.alias);
    return created;
  });
}

function assertTemplateArtifact(
  db: DbFixture,
  templateName: string,
  snapshotID: string,
): void {
  const result = db.queryValue(`
    SELECT revision.worker_spec_snapshot_id, artifact.artifact_digest
    FROM orchestration_resources resource
    JOIN orchestration_resource_revisions revision
      ON revision.organization_id = resource.organization_id
      AND revision.resource_id = resource.id
      AND revision.revision = resource.active_revision
    JOIN worker_spec_dependency_artifacts artifact
      ON artifact.organization_id = resource.organization_id
      AND artifact.worker_spec_snapshot_id = revision.worker_spec_snapshot_id
    WHERE resource.organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    )
      AND resource.kind = 'WorkerTemplate'
      AND resource.name = '${templateName}';
  `);
  const [actualSnapshotID, digest] = result?.split("|") ?? [];
  if (
    actualSnapshotID !== snapshotID ||
    !/^sha256:[0-9a-f]{64}$/.test(digest ?? "")
  ) {
    throw new Error("WorkerTemplate apply did not persist a matching dependency artifact");
  }
}
