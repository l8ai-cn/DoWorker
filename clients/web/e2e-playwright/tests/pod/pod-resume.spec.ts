// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { terminateAllPods } from "../../helpers/pod-cleanup";

type Pod = { podKey: string; status: string };
type ConnectClient = Awaited<ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>>;

test.describe("Pod Resume", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  async function createAndWaitPod(cc: ConnectClient): Promise<string> {
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);
    const resp = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = resp.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();

    await pollUntil(
      async () => {
        const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey: podKey! }) as Pod;
        return pod.status === "running";
      },
      { maxAttempts: 10, intervalMs: 3000, label: "pod-running" },
    ).catch(() => {});

    return podKey!;
  }

  test("terminate and resume pod preserves sandbox", async ({ api }) => {
    const cc = await api.connect();
    const podKey = await createAndWaitPod(cc);

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });

    await pollUntil(
      async () => {
        const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
        return ["terminated", "completed"].includes(pod.status);
      },
      { maxAttempts: 5, intervalMs: 2000, label: "pod-terminated" },
    ).catch(() => {});

    const { agentSlug } = await resolveE2EPodCreateTargets(cc);
    const resumeResp = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      agentSlug,
      sourcePodKey: podKey,
    }) as { pod: Pod };
    const newPodKey = resumeResp.pod?.podKey;
    expect(newPodKey).toBeTruthy();

    if (newPodKey) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: newPodKey });
    }
  });

  test("double resume returns error", async ({ api }) => {
    const cc = await api.connect();
    const podKey = await createAndWaitPod(cc);

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    await new Promise((r) => setTimeout(r, 2000));

    const { agentSlug } = await resolveE2EPodCreateTargets(cc);

    const r1 = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      agentSlug,
      sourcePodKey: podKey,
    }) as { pod: Pod };
    const newKey = r1.pod?.podKey;

    let caught: { status?: number } | null = null;
    try {
      await cc.pod.createPod({
        orgSlug: TEST_ORG_SLUG,
        agentSlug,
        sourcePodKey: podKey,
      });
    } catch (e) {
      caught = e as { status?: number };
    }
    expect(caught).not.toBeNull();
    expect([400, 409]).toContain(caught?.status);

    if (newKey) await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: newKey });
  });
});
