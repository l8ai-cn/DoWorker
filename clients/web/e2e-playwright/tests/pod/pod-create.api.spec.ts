// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { clearAuthRateLimit } from "../../helpers/redis";

type Pod = { podKey: string };

test.describe("Pod Create API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("create basic pod", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);

    const created = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey).toBeTruthy();

    if (podKey) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    }
  });

  test("terminate pod", async ({ api }) => {
    const cc = await api.connect();
    const { runnerId, agentSlug } = await resolveE2EPodCreateTargets(cc);

    const created = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug,
    }) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
  });
});
