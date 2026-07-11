// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { assertNoWasmRecursiveBorrow } from "../../helpers/console-monitor";
import { E2E_ECHO_AGENT_SLUG, pickE2EEchoRunner } from "../../helpers/e2e-echo-runner";

type Runner = { id: bigint };
type Pod = { podKey: string };

/**
 * Regression: opening a pod terminal triggered
 *   "recursive use of an object detected which would lead to unsafe aliasing in rust"
 * because every ACP-aware component synchronously called wasm `get_session_json`
 * during React render, racing the `&mut self` mutators driven by relay events.
 *
 * The fix moved wasm reads inside store mutators (SSOT cache); React render
 * only touches the cache. This spec asserts the page never surfaces that
 * exact wasm-bindgen borrow-check error.
 */
test.describe("ACP terminal: no wasm recursive borrow", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateAllPods(); });

  test("workspace page load does not trigger wasm recursive borrow", async ({ page, monitor }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");
    // Let any deferred ACP/relay subscriptions settle.
    await page.waitForTimeout(1500);

    assertNoWasmRecursiveBorrow(monitor.errors());
  });

  test("creating a pod and rendering its terminal does not trigger wasm recursive borrow", async ({ page, api, monitor }) => {
    const cc = await api.connect();
    const { items: runners } = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG }) as { items: Runner[] };
    expect(runners.length, "dev env must have an online runner").toBeGreaterThan(0);
    const runnerId = pickE2EEchoRunner(runners).id;

    const created = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug: E2E_ECHO_AGENT_SLUG,
    }) as { pod: Pod };
    const podKey = created.pod?.podKey;
    expect(podKey, "pod creation must return a pod_key").toBeTruthy();

    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");

    // Give the relay time to push an ACP snapshot + at least one event so
    // the multi-hook AgentPanel render path actually exercises the cache.
    await page.waitForTimeout(4000);

    assertNoWasmRecursiveBorrow(monitor.errors());
  });
});
