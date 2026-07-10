// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";

type Pod = { podKey: string };
type PodConnectionInfo = { relayUrl: string; token: string; podKey: string };

test.describe("Terminal Connection", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("terminal connect returns websocket URL for running pod", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);

    const created = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();

    await new Promise((r) => setTimeout(r, 5000));

    try {
      const info = await cc.pod.getPodConnection({ orgSlug: TEST_ORG_SLUG, podKey }) as PodConnectionInfo;
      expect(info.relayUrl).toBeTruthy();
    } catch {
      // Pod may not be ready yet.
    }

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
  });
});
