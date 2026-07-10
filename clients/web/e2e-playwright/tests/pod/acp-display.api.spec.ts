// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { clearAuthRateLimit } from "../../helpers/redis";

type Pod = { podKey: string };

test.describe("ACP Pod API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("create ACP pod with agent_slug", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);

    const resp = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = resp.pod?.podKey;
    expect(podKey).toBeTruthy();

    if (podKey) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    }
  });

  test("send prompt to pod via API", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);

    const created = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();

    await new Promise((r) => setTimeout(r, 3000));

    try {
      await cc.pod.sendPodPrompt({
        orgSlug: TEST_ORG_SLUG,
        podKey,
        prompt: "Hello from E2E test",
      });
    } catch (e) {
      const err = e as { status?: number };
      expect([400, 404, 409]).toContain(err.status);
    }

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
  });
});
