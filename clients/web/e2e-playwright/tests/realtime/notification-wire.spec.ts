// Wire-level realtime verification for notification events.
//
// Notification is a *derived* event — the NotificationDispatcher fires
// `proto.events.v1.NotificationPayloadEventData` (target_user_id-scoped)
// in response to other EventBus events like task:completed. We trigger
// it indirectly by terminating a pod and asserting the user-scoped
// notification arrives on the same EventsService.Subscribe stream.
//
// Pod terminate path: cmd/server/eventbus_pod.go:67 → notifDispatcher.Dispatch.
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { withEventSubscription } from "../../helpers/eventbus-stream";
import { createMockAgentPod } from "../../helpers/mock-agent";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";

test.describe("Realtime · notification (wire)", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("notification arrives after task:completed dispatcher fires", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    const { event } = await withEventSubscription<unknown, { source?: string; title?: string; body?: string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "notification" &&
          typeof data.body === "string" && data.body.includes(pod.podKey),
        timeoutMs: 15_000,
      },
      async () => {
        await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey });
      },
    );

    expect(event.type).toBe("notification");
    expect(typeof event.data.title).toBe("string");
    expect(event.data.body).toContain(pod.podKey);
  });
});
