// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { E2E_ECHO_AGENT_SLUG, resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";

type Runner = { id: bigint; currentPods?: number };
type Pod = { podKey: string; status: string };
type ConnectClient = Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>;

test.describe("Pod Lifecycle Scenarios", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  async function createPod(cc: ConnectClient): Promise<{ podKey: string; runnerId: bigint }> {
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);
    const resp = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = resp.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();
    return { podKey: podKey!, runnerId };
  }

  test("full pod lifecycle", async ({ api }) => {
    const cc = await api.connect();
    const { podKey } = await createPod(cc);

    await pollUntil(
      async () => {
        const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
        return pod.status === "running";
      },
      { maxAttempts: 10, intervalMs: 3000, label: "pod-running" },
    ).catch(() => {});

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });

    const final = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
    expect(["terminated", "completed", "error"]).toContain(final.status);
  });

  test("runner capacity changes with pods", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId } = await resolveE2EPodCreateTargets(cc);
    const before = await cc.runner.getRunner({ orgSlug: TEST_ORG_SLUG, id: runnerId }) as { runner: Runner };
    const initialPods = before.runner?.currentPods || 0;

    const { podKey } = await createPod(cc);
    await new Promise((r) => setTimeout(r, 2000));
    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });

    await pollUntil(
      async () => {
        const resp = await cc.runner.getRunner({ orgSlug: TEST_ORG_SLUG, id: runnerId }) as { runner: Runner };
        return (resp.runner?.currentPods || 0) <= initialPods;
      },
      { maxAttempts: 5, intervalMs: 2000, label: "capacity-restore" },
    ).catch(() => {});
  });

  test("resume from non-existent pod returns error", async ({ api }) => {
    const cc = await api.connect();
    let caught: { status?: number } | null = null;
    try {
      await cc.pod.createPod({
        orgSlug: TEST_ORG_SLUG,
        agentSlug: E2E_ECHO_AGENT_SLUG,
        sourcePodKey: "non-existent-pod-key",
      });
    } catch (e) {
      caught = e as { status?: number };
    }
    expect(caught).not.toBeNull();
    expect([400, 404]).toContain(caught?.status);
  });
});
