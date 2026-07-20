// Wire-level realtime EventBus verification for pod:* events.
//
// Asserts each Connect-RPC mutation publishes the expected typed
// `proto.events.v1.*EventData` payload (see proto/events/v1/event_data.proto)
// onto the EventsService.Subscribe stream that the production renderer
// consumes. Failures here mean either:
//   - backend stopped publishing the event after the mutation
//   - proto field names drifted (UseProtoNames=true keeps snake_case wire)
//   - event_data schema changed without consumer-side update
//
// Coverage: pod:created / pod:terminated / pod:alias_changed /
// pod:perpetual_changed / pod:status_changed (B-class, observed during
// mock-agent lifecycle).
//
// UI propagation is exercised separately in tests/pod/pod-events-multitab.spec.ts.
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { withEventSubscription } from "../../helpers/eventbus-stream";
import { createMockAgentPod } from "../../helpers/mock-agent";
import {
  createE2EPodAlias,
  terminateRegisteredE2EPods,
  updateE2ECreatedPodAlias,
} from "../../helpers/pod-cleanup";
import { createE2EEchoPod } from "../../helpers/e2e-worker-spec";

test.describe("Realtime · pod events (wire)", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("pod:created arrives with pod_key + runner_id + ticket fields", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");
    const ticket = await cc.ticket.createTicket({
      orgSlug: TEST_ORG_SLUG,
      title: `E2E pod event ${Date.now()}`,
    }) as { id: bigint; slug: string };
    try {
      const { event, actionResult } = await withEventSubscription<
        { pod?: { podKey: string; runnerId: bigint } },
        { pod_key?: string; runner_id?: number | string; ticket_slug?: string }
      >(
        {
          token, orgSlug: TEST_ORG_SLUG,
          predicate: (type, data) =>
            type === "pod:created" &&
            Number(data.ticket_id) === Number(ticket.id),
        },
        () => createE2EEchoPod(cc, {
          ticketSlug: ticket.slug,
        }) as Promise<{ pod?: { podKey: string; runnerId: bigint } }>,
      );
      const pod = actionResult.pod;
      expect(pod?.podKey).toBeTruthy();
      expect(event.data.pod_key).toBe(pod?.podKey);
      expect(Number(event.data.runner_id)).toBe(Number(pod?.runnerId));
      expect(Number(event.data.ticket_id)).toBe(Number(ticket.id));
      expect(event.data.ticket_slug).toBe(ticket.slug);
    } finally {
      await cc.ticket.deleteTicket({
        orgSlug: TEST_ORG_SLUG,
        ticketSlug: ticket.slug,
      }).catch(() => undefined);
    }
  });

  test("pod:terminated arrives with terminal status", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    const { event } = await withEventSubscription<unknown, { pod_key?: string; status?: string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "pod:terminated" && data.pod_key === pod.podKey,
      },
      async () => {
        await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey });
      },
    );

    expect(event.data.pod_key).toBe(pod.podKey);
    expect(typeof event.data.status).toBe("string");
    expect(["terminated", "completed", "error"]).toContain(event.data.status);
  });

  test("pod:alias_changed arrives with new alias", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });
    const newAlias = createE2EPodAlias(`alias-${Date.now().toString(36)}`);

    const { event } = await withEventSubscription<unknown, { pod_key?: string; alias?: string }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "pod:alias_changed" && data.pod_key === pod.podKey,
      },
      async () => {
        await cc.pod.updatePodAlias({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey, alias: newAlias });
        updateE2ECreatedPodAlias(pod.podKey, newAlias);
      },
    );

    expect(event.data.pod_key).toBe(pod.podKey);
    expect(event.data.alias).toBe(newAlias);
  });

  test("pod:perpetual_changed arrives with perpetual=true", async ({ api }) => {
    const cc = await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    const { event } = await withEventSubscription<unknown, { pod_key?: string; perpetual?: boolean }>(
      {
        token, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "pod:perpetual_changed" && data.pod_key === pod.podKey,
      },
      async () => {
        await cc.pod.updatePodPerpetual({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey, perpetual: true });
      },
    );

    expect(event.data.pod_key).toBe(pod.podKey);
    expect(event.data.perpetual).toBe(true);
  });

  test("pod:status_changed fires during pod lifecycle", async ({ api }) => {
    await api.connect();
    const token = api.getToken();
    if (!token) throw new Error("api fixture missing token");

    // status_changed is B-class (runner-driven) — we don't directly
    // trigger it, just observe that creating a pod produces at least one
    // pod:status_changed within the 10s window. Mock-agent transitions
    // initializing → running near-immediately.
    let createdPodKey: string | undefined;
    const { event } = await withEventSubscription<unknown, { pod_key?: string; status?: string }>(
      {
        token: token!, orgSlug: TEST_ORG_SLUG,
        predicate: (type, data) =>
          type === "pod:status_changed" && data.pod_key === createdPodKey,
        timeoutMs: 15_000,
      },
      async () => {
        const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });
        createdPodKey = pod.podKey;
      },
    );

    expect(event.data.pod_key).toBe(createdPodKey);
    expect(typeof event.data.status).toBe("string");
  });
});
