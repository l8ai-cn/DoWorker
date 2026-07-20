// Wire-level realtime EventBus verification for autopilot:created.
//
// Other autopilot events (status_changed / iteration / thinking /
// terminated) are runner-driven via the autopilot loop in the pod's
// runner process — listed as Out of Scope, covered by backend
// integration tests (see plan).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { withEventSubscription } from "../../helpers/eventbus-stream";
import { createMockAgentPod } from "../../helpers/mock-agent";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";

test.describe("Realtime · autopilot events (wire)", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("autopilot:created arrives with autopilot_controller_key + pod_key", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    const { event } = await withEventSubscription<unknown, { autopilot_controller_key?: string; pod_key?: string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "autopilot:created" && data.pod_key === pod.podKey,
      },
      async () => {
        await cc.autopilot.createAutopilotController({
          orgSlug: TEST_ORG_SLUG, podKey: pod.podKey, prompt: "verify autopilot bridge",
        } as never);
      },
    );

    expect(event.data.pod_key).toBe(pod.podKey);
    expect(typeof event.data.autopilot_controller_key).toBe("string");
    expect(event.data.autopilot_controller_key!.length).toBeGreaterThan(0);
  });
});
