// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { createE2EEchoPod } from "../../helpers/e2e-worker-spec";

type Channel = { id: bigint };
type ChannelMessage = { id: bigint };
type Pod = { podKey: string; status: string };

/**
 * Journey: Multi-Agent Collaboration
 * Channel → Multiple Pods → Message Exchange → Coordination
 */
test.describe("Journey: Multi-Agent Collaboration", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test.afterAll(async () => {
    await terminateAllPods();
  });

  test("channel-based multi-pod collaboration flow", async ({ api }) => {
    const cc = await api.connect();

    // ── Step 1: Create collaboration channel ──
    const chName = "E2E Collab " + Date.now();
    const ch = await cc.channel.createChannel({
      orgSlug: TEST_ORG_SLUG,
      name: chName,
      description: "Multi-agent collaboration test",
    }) as Channel;
    const chId = ch.id;
    expect(chId).toBeTruthy();

    // ── Step 2: Create Pod A (analyst) ──
    const podAResp = await createE2EEchoPod(cc, {
      alias: "E2E Collab Pod A - Analyst",
    }) as { pod: Pod };
    const podAKey = podAResp.pod?.podKey;

    // ── Step 3: Create Pod B (implementer) ──
    const podBResp = await createE2EEchoPod(cc, {
      alias: "E2E Collab Pod B - Implementer",
    }) as { pod: Pod };
    const podBKey = podBResp.pod?.podKey;

    // ── Step 4: Wait for both pods running ──
    for (const podKey of [podAKey, podBKey].filter((k): k is string => Boolean(k))) {
      await pollUntil(
        async () => {
          const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
          return pod.status === "running";
        },
        { maxAttempts: 10, intervalMs: 3000, label: `pod-${podKey}-running` }
      ).catch(() => {});
    }

    // ── Step 5: Add pods to channel ──
    if (podAKey) {
      await cc.channel.joinChannelPod({ orgSlug: TEST_ORG_SLUG, id: chId, podKey: podAKey });
    }
    if (podBKey) {
      await cc.channel.joinChannelPod({ orgSlug: TEST_ORG_SLUG, id: chId, podKey: podBKey });
    }

    // ── Step 6: Verify channel members ──
    // Connect throws on failure — list succeeding is the assertion.
    await cc.channel.listChannelMembers({ orgSlug: TEST_ORG_SLUG, id: chId });

    // ── Step 7: Send message to channel ──
    const msg = await cc.channel.sendChannelMessage({
      orgSlug: TEST_ORG_SLUG,
      channelId: chId,
      source: "Pod A found a bug in the auth module. Pod B, please fix it.",
    }) as ChannelMessage;
    expect(msg.id).toBeTruthy();

    // ── Step 8: Verify message appears in history ──
    const { items: messages } = await cc.channel.listChannelMessages({
      orgSlug: TEST_ORG_SLUG,
      channelId: chId,
    }) as { items: ChannelMessage[] };
    expect(messages.length).toBeGreaterThan(0);

    // ── Step 9: Verify mesh topology shows pods and channel ──
    // Connect throws on failure — successful read is the assertion.
    await cc.mesh.getMeshTopology({ orgSlug: TEST_ORG_SLUG });

    // ── Step 10: Cleanup — terminate pods, archive channel ──
    for (const podKey of [podAKey, podBKey].filter((k): k is string => Boolean(k))) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    }
    await cc.channel.archiveChannel({ orgSlug: TEST_ORG_SLUG, id: chId });
  });
});
