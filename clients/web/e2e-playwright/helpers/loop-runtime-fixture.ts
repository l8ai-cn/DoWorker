import type { DbFixture } from "../fixtures/db.fixture";
import type { ApiFixture } from "../fixtures/api.fixture";
import { createE2EEchoPod } from "./e2e-worker-spec";
import { TEST_ORG_SLUG } from "./env";
import { unregisterE2ECreatedPod } from "./pod-cleanup";

export interface LoopRuntimeFixture {
  alias: string;
  goalLoopName: string;
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
  const created = await createE2EEchoPod(client, {
    alias: `Loop runtime ${suffix}`,
    automationLevel: "autonomous",
  });
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
  await client.pod.terminatePod({
    orgSlug: TEST_ORG_SLUG,
    podKey: sourcePodKey,
  });
  unregisterE2ECreatedPod(sourcePodKey);
  return {
    alias,
    goalLoopName: `loop-runtime-${suffix}`,
    snapshotId,
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
