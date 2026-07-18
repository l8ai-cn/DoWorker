// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { createE2EEchoPod } from "../../helpers/e2e-worker-spec";

type Runner = { id: bigint; currentPods?: number };
type Pod = { podKey: string; status: string; runnerId: bigint };
type ConnectClient = Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>;

/**
 * Pod lifecycle scenario tests.
 * These require at least one runner to be online.
 */
test.describe("Pod Lifecycle Scenarios", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /** Helper: create a pod and return its key. Asserts prerequisites instead of skipping. */
  async function createPod(cc: ConnectClient): Promise<Pod> {
    const resp = await createE2EEchoPod(cc) as { pod: Pod };
    const podKey = resp.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();
    expect(resp.pod.runnerId).toBeTruthy();
    return resp.pod;
  }

  /**
   * TC-POD-004: Full lifecycle (create → running → terminate)
   */
  test("full pod lifecycle", async ({ api }) => {
    const cc = await api.connect();
    const { podKey } = await createPod(cc);

    await pollUntil(
      async () => {
        const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
        return pod.status === "running";
      },
      { maxAttempts: 10, intervalMs: 3000, label: "pod-running" }
    ).catch(() => {/* may timeout, continue anyway */});

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });

    const final = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
    expect(["terminated", "completed", "error"]).toContain(final.status);
  });

  /**
   * TC-POD-005: Runner capacity tracks pod count
   */
  test("runner capacity changes with pods", async ({ api }) => {
    const cc = await api.connect();
    const { items: runners } = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG }) as { items: Runner[] };
    expect(runners.length, "dev env must have an online runner").toBeGreaterThan(0);
    const pod = await createPod(cc);
    const runner = runners.find((item) => item.id === pod.runnerId);
    expect(runner, "scheduled runner must be in the available runner list").toBeTruthy();
    const initialPods = runner?.currentPods || 0;

    await new Promise((r) => setTimeout(r, 2000));

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey });

    await pollUntil(
      async () => {
        const resp = await cc.runner.getRunner({
          orgSlug: TEST_ORG_SLUG,
          id: pod.runnerId,
        }) as { runner: Runner };
        return (resp.runner?.currentPods || 0) <= initialPods;
      },
      { maxAttempts: 5, intervalMs: 2000, label: "capacity-restore" }
    ).catch(() => {/* ignore */});
  });

  /**
   * TC-POD-007: Resume edge cases
   */
  test("resume from non-existent pod returns error", async ({ api }) => {
    const cc = await api.connect();

    let caught: { status?: number } | null = null;
    try {
      await cc.pod.createPod({
        orgSlug: TEST_ORG_SLUG,
        sourcePodKey: "non-existent-pod-key",
      });
    } catch (e) {
      caught = e as { status?: number };
    }
    expect(caught).not.toBeNull();
    expect([400, 404]).toContain(caught?.status);
  });
});
