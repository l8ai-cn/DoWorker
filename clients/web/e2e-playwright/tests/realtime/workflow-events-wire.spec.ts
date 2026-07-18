// Wire-level realtime EventBus verification for workflow_run:* events.
//
// Covers workflow_run:started (A-class — direct from TriggerWorkflow), plus
// workflow_run:completed / workflow_run:failed (B-class — via pod terminate +
// orchestrator listener chain in backend/internal/service/workflow/).
//
// workflow_run:warning is not directly triggerable from e2e (requires a
// timing failure inside the orchestrator); listed as Out of Scope in
// the plan and to be covered by backend integration test.
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { withEventSubscription, subscribeEvents } from "../../helpers/eventbus-stream";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { createResourceWorkflow } from "../../helpers/resource-workflow";

test.describe("Realtime · workflow_run events (wire)", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateAllPods(); });

  async function createWorkflow(api: import("../../fixtures/api.fixture").ApiFixture) {
    const cc = await api.connect();
    const stamp = Date.now().toString(36);
    return createResourceWorkflow(cc, {
      name: `e2e-rt-workflow-${stamp}`,
      slug: `e2e-rt-workflow-${stamp}`,
      prompt: "echo hi",
      executionMode: "direct",
      timeoutMinutes: 1,
    });
  }

  test("workflow_run:started arrives after TriggerWorkflow", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");
    const workflow = await createWorkflow(api);

    const { event } = await withEventSubscription<unknown, { workflow_id?: number | string; run_id?: number | string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type) => type === "workflow_run:started",
        timeoutMs: 15_000,
      },
      async () => {
        await cc.workflow.triggerWorkflow({
          orgSlug: TEST_ORG_SLUG, workflowSlug: workflow.slug,
        } as never);
      },
    );

    expect(event.type).toBe("workflow_run:started");
    expect(Number(event.data.workflow_id)).toBeGreaterThan(0);
    expect(Number(event.data.run_id)).toBeGreaterThan(0);
  });

  test("workflow_run:completed arrives after pod successful termination", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");
    const workflow = await createWorkflow(api);

    // Backend publishes workflow_run:started twice: once at TriggerWorkflow time
    // (before pod creation; pod_key empty) and once after SetRunPodKey
    // commits (pod_key populated). We wait for the second emission so we
    // can correlate the run to its pod for termination.
    const ctrl = new AbortController();
    let runPodKey: string | null = null;
    let runId: number | null = null;

    const startedListener = (async () => {
      for await (const ev of subscribeEvents({ token, orgSlug: TEST_ORG_SLUG, signal: ctrl.signal })) {
        if (ev.type !== "workflow_run:started") continue;
        const data = JSON.parse(ev.dataJson) as { pod_key?: string; run_id?: number | string };
        if (data.pod_key && data.pod_key.length > 0) {
          runPodKey = data.pod_key;
          runId = Number(data.run_id);
          ctrl.abort();
          return;
        }
      }
    })();

    await new Promise((r) => setTimeout(r, 500));
    await cc.workflow.triggerWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: workflow.slug } as never);

    const startedRace = await Promise.race([
      startedListener.then(() => "got-pod" as const),
      new Promise<"timeout">((r) => setTimeout(() => r("timeout"), 20_000)),
    ]);
    ctrl.abort();
    expect(startedRace, "workflow_run:started with pod_key must arrive within 20s").toBe("got-pod");
    expect(runPodKey, "pod_key must be bound to the run").toBeTruthy();

    const { event } = await withEventSubscription<unknown, { run_id?: number | string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          (type === "workflow_run:completed" || type === "workflow_run:failed") &&
          Number(data.run_id) === runId,
        timeoutMs: 20_000,
      },
      async () => {
        await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: runPodKey! });
      },
    );

    expect(Number(event.data.run_id)).toBe(runId);
  });
});
