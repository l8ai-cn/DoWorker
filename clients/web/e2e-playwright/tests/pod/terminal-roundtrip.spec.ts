import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { createMockAgentPod, workspaceUrlForPod } from "../../helpers/mock-agent";
import { takeWorkerControl } from "../../helpers/acp-spec-setup";

// End-to-end coverage of the terminal DATA PLANE after the relay-SSOT
// migration: browser xterm ↔ relayConnection adapter ↔ WasmRelayManager ↔
// Rust RelayConnectionPool ↔ relay WS ↔ runner PTY ↔ e2e-echo (pty mode).
//
// This is the only spec that exercises the relay OUTPUT byte path through the
// adapter (acp-ui-echo covers the ACP-message path). The Rust pool owns
// reconnect/dedup/debounce/codec/snapshot replay; the surviving JS adapter
// must still wire real bytes both ways, verified by the echo round-trip: a
// typed line travels IN (xterm → adapter → … → PTY) and the PTY's reply
// travels back OUT (PTY → … → adapter → xterm).
// pty_runtime.go: writes "ready" on spawn, then echoes each stdin line as
// "got: <line>". We assert the echo (delivered live once subscribed), not the
// one-shot spawn banner — see the round-trip note in the test body.
test.describe("Terminal data-plane round-trip (relay SSOT)", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("attaches, streams pty output, and round-trips typed input through the relay", async ({ page, api, monitor }) => {
    // Realtime EventsService streams through the Next dev-server proxy in local
    // e2e; that proxy intermittently 502s long-lived gRPC streams. It only
    // affects the control-plane event feed (pod-status push), never the relay
    // data plane this spec asserts — acp-ui-echo proves the relay path over the
    // same adapter. Wait-for-running below removes the dependency on the event.
    monitor.allow(/EventsService\/Subscribe.*502|Subscribe:0:0.*502/);

    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    // Gate navigation on the pod being relay-ready (getPodConnection only
    // returns a relay URL once the pod is running). The workspace's initial
    // pod fetch then sees "running" → AgentPanel subscribes the terminal
    // without waiting for the (proxy-flaky) realtime pod-status event.
    const cc = await api.connect();
    await expect(async () => {
      const info = await cc.pod.getPodConnection({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey }) as { relayUrl?: string };
      expect(info.relayUrl, "pod not relay-ready yet").toBeTruthy();
    }).toPass({ timeout: 30_000 });

    await page.goto(workspaceUrlForPod(pod.podKey));
    await page.waitForLoadState("load");
    await takeWorkerControl(page);

    // xterm uses the DOM renderer (fit/weblinks/search addons only — no
    // webgl/canvas), so rendered rows are queryable text. Wait for the
    // terminal to mount + its hidden input to attach before driving I/O.
    const term = page.locator(".xterm");
    await expect(term).toBeVisible({ timeout: 30_000 });
    const input = page.locator(".xterm-helper-textarea");
    await expect(input).toBeAttached({ timeout: 30_000 });
    await takeWorkerControl(page);

    // Assert the round-trip on the INPUT echo, delivered LIVE once the browser
    // subscribes: typed line → onData → relayPool.send → WasmRelayManager →
    // Rust pool → relay → runner PTY → "got: <line>" → back through the relay
    // → xterm. This exercises the surviving adapter's byte path BOTH ways.
    //
    // We deliberately do NOT gate on the one-shot "ready" spawn banner: it is
    // written before the browser subscribes, so catching it E2E hinges on the
    // runner's early-output replay landing ahead of the relay snapshot — a race
    // covered at the daemon layer by TestEarlyOutputReplayedOnAttach and too
    // timing-sensitive to gate this live-infra spec on. Retrying type+assert
    // absorbs the subscription-establishment window (input typed before the
    // subscription is wired is dropped at the PTY, not buffered).
    await expect(async () => {
      await input.focus();
      await page.keyboard.type("relay-roundtrip");
      await page.keyboard.press("Enter");
      await expect(term, "pty echo of typed input must round-trip back to xterm").toContainText(
        "got: relay-roundtrip",
        { timeout: 8_000 },
      );
    }).toPass({ timeout: 60_000 });
  });
});
