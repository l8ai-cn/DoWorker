import { expect, type Page } from "@playwright/test";
import type { ApiFixture } from "../fixtures/api.fixture";
import {
  assertNoWasmRecursiveBorrow,
  type ConsoleMonitor,
} from "./console-monitor";
import {
  createMockAgentPod,
  workspaceUrlForPod,
  type CreateMockPodOptions,
  type MockAgentPod,
} from "./mock-agent";

// SetupAcpScenarioResult bundles the per-spec context so each test can
// reach for exactly what it needs without a fixture explosion.
export interface SetupAcpScenarioResult {
  pod: MockAgentPod;
  /** Asserts the wasm-bindgen recursive-borrow guard is intact. */
  assertWasmHealthy: () => void;
}

export async function takeWorkerControl(page: Page): Promise<void> {
  const takeControl = page.getByRole("button", { name: "Take control" });
  await expect(takeControl).toBeEnabled({ timeout: 30_000 });
  await takeControl.click();
  await expect(takeControl).toBeHidden({ timeout: 30_000 });
}

// setupAcpScenarioPage encapsulates the 4-step prologue every ACP UI spec
// repeats: spawn mock pod → navigate to the workspace → wait for load →
// expose `assertWasmHealthy` so the spec can check the borrow guard at
// any point. Throws when no runner is online — the e2e suite contract is
// "dev env has at least one online runner", so returning null here would
// silently mask a missing prerequisite.
//
// Console/page error collection is owned by the auto-attached `monitor`
// fixture (see fixtures/index.ts) — callers pass it in so this helper
// doesn't re-attach listeners and so teardown's default-deny assert sees
// the same buffer the recursive-borrow guard inspects.
export async function setupAcpScenarioPage(
  page: Page,
  api: ApiFixture,
  monitor: ConsoleMonitor,
  opts: CreateMockPodOptions,
): Promise<SetupAcpScenarioResult> {
  const pod = await createMockAgentPod(api, opts);
  if (!pod) {
    throw new Error("setupAcpScenarioPage: createMockAgentPod returned null — dev env must have an online runner");
  }

  await page.goto(workspaceUrlForPod(pod.podKey));
  // Connect-RPC EventsService streams keep the page in flight indefinitely
  // on r6, so "networkidle" (the original pre-r6 strategy) times out. "load"
  // matches what pod-create-ui.spec.ts and pod-lifecycle.spec.ts use against
  // the same workspace route.
  await page.waitForLoadState("load");

  return {
    pod,
    assertWasmHealthy: () => {
      assertNoWasmRecursiveBorrow(monitor.errors());
    },
  };
}
