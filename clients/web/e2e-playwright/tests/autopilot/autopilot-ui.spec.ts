import { test, expect } from "../../fixtures/index";
import type { Page } from "@playwright/test";
import type { ApiFixture } from "../../fixtures/api.fixture";
import type { ConsoleMonitor } from "../../helpers/console-monitor";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { workspaceUrlForPod } from "../../helpers/mock-agent";
import { createReadyAutopilotTarget, createAutopilotForPod } from "../../helpers/autopilot";

// Opens the workspace, attaches a running controller after the page is live on
// realtime (the autopilot:created edge is what hydrates the store), and returns
// the visible status bar locator.
async function openWorkspaceWithAutopilot(page: Page, api: ApiFixture, monitor: ConsoleMonitor) {
  monitor.allow(/Failed to load resource.*50\d \(/); // transient realtime gateway hiccup
  await api.login();
  const pod = await createReadyAutopilotTarget(api);

  await page.goto(workspaceUrlForPod(pod.podKey));
  await page.waitForLoadState("load");
  await page.waitForTimeout(3000);
  const takeControl = page.getByRole("button", { name: "Take control" });
  await expect(takeControl).toBeEnabled({ timeout: 30_000 });
  await takeControl.click();
  await expect(takeControl).toBeHidden({ timeout: 30_000 });

  await createAutopilotForPod(api, {
    targetPodKey: pod.podKey,
    script: { decisions: [{ type: "continue", reasoning: "loop", send_input: "echo loop\n" }] },
    maxIterations: 20,
  });

  const statusBar = page.getByTestId("autopilot-status-bar");
  await expect(statusBar).toBeVisible({ timeout: 30_000 });
  return statusBar;
}

test.describe("Autopilot UI · status bar", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  test.afterEach(async () => {
    await terminateRegisteredE2EPods();
  });

  test("workspace renders the autopilot status bar with the live phase", async ({ page, api, monitor }) => {
    test.setTimeout(120_000);
    const statusBar = await openWorkspaceWithAutopilot(page, api, monitor);
    await expect(statusBar).toHaveAttribute(
      "data-phase",
      /initializing|running|paused|waiting_approval|user_takeover|completed|max_iterations/,
    );
  });

  test("status bar pause/resume buttons drive the controller", async ({ page, api, monitor }) => {
    test.setTimeout(120_000);
    const statusBar = await openWorkspaceWithAutopilot(page, api, monitor);
    await expect(statusBar).toHaveAttribute("data-phase", "running", { timeout: 30_000 });

    // Button click → control RPC → runner → status_changed → store → DOM.
    await statusBar.getByTitle("Pause").click();
    await expect(statusBar).toHaveAttribute("data-phase", "paused", { timeout: 20_000 });

    await statusBar.getByTitle("Resume").click();
    await expect(statusBar).toHaveAttribute("data-phase", "running", { timeout: 20_000 });
  });

  test("view-details opens the autopilot bottom panel", async ({ page, api, monitor }) => {
    test.setTimeout(120_000);
    const statusBar = await openWorkspaceWithAutopilot(page, api, monitor);
    await statusBar.getByTitle("View Details").click();
    await expect(page.getByTestId("autopilot-panel")).toBeVisible({ timeout: 15_000 });
  });
});
