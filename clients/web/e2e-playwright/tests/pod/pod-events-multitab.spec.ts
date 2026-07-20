// Multi-tab UI propagation for pod lifecycle events.
//
// Validates the full wire→UI chain that wire-level specs cannot:
//   Connect-RPC mutation (tab A)
//     → backend EventBus.Publish
//     → EventsService.Subscribe stream (tab B's subscriber)
//     → wasm EventSubscriptionManager → handler → Zustand store
//     → React re-render → sidebar pod-list-item updates
//
// Failure here means one of: realtime event dropped, handler regression,
// store mutation broken, or React subscription bug — none of which the
// wire-level pod-events-wire.spec.ts can catch.
import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { createMockAgentPod } from "../../helpers/mock-agent";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Pod events · multi-tab UI propagation", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("tab A terminate → tab B sidebar removes the pod", async ({ context, api }) => {
    const cc = await api.connect();
    const pod = await createMockAgentPod(api, { mode: "pty", scenario: "echo" });

    // Both tabs land on the workspace page so the sidebar pod list is mounted
    // and EventSubscriptionManager has had time to bootstrap its Connect
    // server-stream subscription before we publish.
    const tabA = await context.newPage();
    const tabB = await context.newPage();
    await Promise.all([
      tabA.goto(`/${TEST_ORG_SLUG}/workspace`),
      tabB.goto(`/${TEST_ORG_SLUG}/workspace`),
    ]);
    await Promise.all([tabA.waitForLoadState("load"), tabB.waitForLoadState("load")]);

    // Both tabs must show the just-created pod in the sidebar before we
    // can validate cross-tab removal. data-pod-key is stable across the
    // sidebar list components (PodListItem.tsx).
    const podSelector = `[data-testid="pod-list-item"][data-pod-key="${pod.podKey}"]`;
    await Promise.all([
      expect(tabA.locator(podSelector)).toBeVisible({ timeout: 15_000 }),
      expect(tabB.locator(podSelector)).toBeVisible({ timeout: 15_000 }),
    ]);

    // EventSubscriptionManager.connect() resolves async after the wasm
    // bootstrap — give the SubscribeRequest a moment to register with the
    // backend hub before publishing. Same pattern as blockstore multi-tab-sync.
    await tabA.waitForTimeout(1500);

    await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: pod.podKey });

    // The terminated pod can either get its status badge flipped or be
    // filtered out of the sidebar list entirely, depending on the active
    // sidebar filter. Either way the pod-list-item with the badge
    // "running" / "initializing" must no longer be visible. Wait until
    // both tabs no longer show the running pod entry.
    await Promise.all([
      expect(tabA.locator(podSelector)).toHaveCount(0, { timeout: 10_000 }),
      expect(tabB.locator(podSelector)).toHaveCount(0, { timeout: 10_000 }),
    ]);

    await tabA.close();
    await tabB.close();
  });
});
