// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { createE2EEchoPod } from "../../helpers/e2e-worker-spec";

type Pod = { podKey: string };

test.describe("Pod Create API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-POD-001: Create basic pod
   */
  test("create basic pod", async ({ api }) => {
    const cc = await api.connect();
    const created = await createE2EEchoPod(cc) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey).toBeTruthy();

    if (podKey) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    }
  });

  /**
   * TC-POD-003: Terminate pod
   */
  test("terminate pod", async ({ api }) => {
    const cc = await api.connect();
    const created = await createE2EEchoPod(cc) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey, "createPod must return a pod_key").toBeTruthy();

    // Connect throws on failure — no need to assert status.
    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
  });
});
