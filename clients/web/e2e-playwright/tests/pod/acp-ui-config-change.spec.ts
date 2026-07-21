import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import {
  createMockAgentPod,
  workspaceUrlForPod,
} from "../../helpers/mock-agent";
import { takeWorkerControl } from "../../helpers/worker-control-lease";

// The shared AgentWorkspace only exposes configuration controls advertised by
// the ACP session. This spec verifies that the advertised permission modes are
// rendered and that a selected mode returns through the control plane.
test.describe("ACP UI: control plane round-trip", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("clicking an advertised permission mode updates the rendered label after server ack", async ({ page, api }) => {
    const pod = await createMockAgentPod(api, {
      mode: "acp",
      scenario: "permission_modes_loopal",
      prompt: "ready",
    });

    await page.goto(workspaceUrlForPod(pod.podKey));
    // Use "load" (matches setupAcpScenarioPage in helpers/acp-spec-setup.ts):
    // Connect-RPC EventsService keeps connections open indefinitely so
    // "networkidle" times out under r6.
    await page.waitForLoadState("load");
    // Wait for the initial acknowledgment chunk so we know wasm session
    // is wired and Selector is mounted.
    await expect(page.getByText("Ready for mode switches", { exact: false })).toBeVisible({ timeout: 15_000 });
    await takeWorkerControl(page);

    const permissions = page.getByRole("combobox", { name: "Permissions" });
    await expect(permissions).toBeVisible();
    await permissions.click();
    await page.getByRole("option", { name: "Full access" }).click();
    await expect(permissions).toHaveText("Full access", { timeout: 10_000 });
  });

  test("loopal-advertised modes render in the selector dropdown (capability path)", async ({ page, api }) => {
    // Mock advertises agentcloudExtensions.permissionModes (loopal's 3 modes);
    // exercises the full advertise → parse → snapshot → selector render path.
    const pod = await createMockAgentPod(api, {
      mode: "acp",
      scenario: "permission_modes_loopal",
      prompt: "ready",
    });

    await page.goto(workspaceUrlForPod(pod.podKey));
    await page.waitForLoadState("load");
    await expect(page.getByText("Ready for mode switches", { exact: false })).toBeVisible({ timeout: 15_000 });
    await takeWorkerControl(page);

    await page.getByRole("combobox", { name: "Permissions" }).click();

    await expect(page.getByRole("option", { name: "Full access" })).toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole("option", { name: "Ask for dangerous actions" })).toBeVisible();
    await expect(page.getByRole("option", { name: "Ask before writes" })).toBeVisible();
    await expect(page.getByRole("option", { name: "Accept edits" })).toHaveCount(0);
  });
});
