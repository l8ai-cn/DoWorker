import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { createResourceWorkflow } from "../../helpers/resource-workflow";

test.describe("Workflow resource EnvBundle binding", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  test("pins ordered EnvBundle IDs in the WorkerSpec snapshot", async ({
    api,
    db,
  }) => {
    const cc = await api.connect();
    const ts = Date.now();
    const bundleA = await createBundle(cc, `e2e-workflow-A-${ts}`);
    const bundleB = await createBundle(cc, `e2e-workflow-B-${ts}`);
    const slug = `e2e-workflow-bundle-${ts}`;
    try {
      const created = await createResourceWorkflow(cc, {
        slug,
        name: `E2E Workflow Bundle ${ts}`,
        prompt: "echo bound",
        environmentBundles: [bundleA, bundleB],
      });
      const fetched = await cc.workflow.getWorkflow({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: slug,
      }) as { slug: string; workerSpecSnapshotId: bigint };
      expect(fetched.slug).toBe(slug);
      expect(fetched.workerSpecSnapshotId).toBe(created.workerSpecSnapshotId);
      expect(snapshotEnvBundleIds(db, created.workerSpecSnapshotId)).toEqual([
        Number(bundleA.id),
        Number(bundleB.id),
      ]);
    } finally {
      await deleteWorkflow(cc, slug);
      await deleteBundle(cc, bundleA.id);
      await deleteBundle(cc, bundleB.id);
    }
  });

  test("legacy Workflow update cannot mutate the pinned snapshot", async ({
    api,
    db,
  }) => {
    const cc = await api.connect();
    const ts = Date.now();
    const bundle = await createBundle(cc, `e2e-clear-${ts}`);
    const slug = `e2e-workflow-clear-${ts}`;
    try {
      const created = await createResourceWorkflow(cc, {
        slug,
        name: `E2E Workflow Clear ${ts}`,
        prompt: "echo bound",
        environmentBundles: [bundle],
      });
      const before = snapshotEnvBundleIds(db, created.workerSpecSnapshotId);
      await cc.workflow.updateWorkflow({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: slug,
        usedEnvBundles: { names: [] },
      });
      expect(snapshotEnvBundleIds(db, created.workerSpecSnapshotId)).toEqual(
        before,
      );
    } finally {
      await deleteWorkflow(cc, slug);
      await deleteBundle(cc, bundle.id);
    }
  });

  test("rejects a dangling EnvBundle resource during plan", async ({ api }) => {
    const cc = await api.connect();
    const ts = Date.now();
    await expect(createResourceWorkflow(cc, {
      slug: `e2e-workflow-dangling-${ts}`,
      name: `E2E Workflow Dangling ${ts}`,
      prompt: "echo dangling",
      environmentBundles: [{
        id: 999_999_999n,
        name: `nonexistent-bundle-${ts}`,
      }],
    })).rejects.toThrow(/EnvBundle|environment bundle|not found|blocked/i);
  });
});

async function createBundle(
  client: Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>,
  name: string,
) {
  return client.envBundle.createEnvBundle({
    agentSlug: "e2e-echo",
    name,
    kind: "runtime",
    data: { E2E_WORKFLOW_VALUE: name },
  }) as Promise<{ id: bigint; name: string }>;
}

function snapshotEnvBundleIds(
  db: import("../../fixtures/db.fixture").DbFixture,
  snapshotId: bigint,
): number[] {
  const raw = db.queryValue(`
    SELECT spec_json #>> '{workspace,env_bundle_ids}'
    FROM worker_spec_snapshots
    WHERE id = ${snapshotId}
  `);
  if (!raw) throw new Error(`WorkerSpec snapshot ${snapshotId} is missing`);
  return (JSON.parse(raw) as number[]).map(Number);
}

async function deleteWorkflow(
  client: Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>,
  slug: string,
) {
  await client.workflow.deleteWorkflow({
    orgSlug: TEST_ORG_SLUG,
    workflowSlug: slug,
  }).catch(() => undefined);
}

async function deleteBundle(
  client: Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>,
  id: bigint,
) {
  await client.envBundle.deleteEnvBundle({ id }).catch(() => undefined);
}
