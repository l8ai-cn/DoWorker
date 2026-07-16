import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { takeWorkerControl } from "../../helpers/acp-spec-setup";
import {
  createMockAgentPod,
  workspaceUrlForPod,
} from "../../helpers/mock-agent";

// Multi-tab synchronization regression for the shared AgentWorkspace
// configuration bar. A mode change in tab A must propagate to tab B without
// a manual refresh.
test.describe("ACP UI: multi-tab Selector synchronization", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateAllPods(); });

  test("mode change in tab A appears in tab B without refresh", async ({ context, api }) => {
    const pod = await createMockAgentPod(api, {
      mode: "acp",
      scenario: "permission_modes_loopal",
      prompt: "multi-tab probe",
    });

    const tabA = await context.newPage();
    const tabB = await context.newPage();

    await Promise.all([
      tabA.goto(workspaceUrlForPod(pod.podKey)),
      tabB.goto(workspaceUrlForPod(pod.podKey)),
    ]);
    // Use "load" — see acp-ui-config-change.spec.ts header for the same r6
    // Connect-RPC streaming rationale.
    await Promise.all([
      tabA.waitForLoadState("load"),
      tabB.waitForLoadState("load"),
    ]);

    // Wait for both tabs to render the initial activity (so both have
    // an active relay subscription and a mounted Selector).
    await Promise.all([
      expect(tabA.getByText("Ready for mode switches", { exact: false })).toBeVisible({ timeout: 15_000 }),
      expect(tabB.getByText("Ready for mode switches", { exact: false })).toBeVisible({ timeout: 15_000 }),
    ]);
    await takeWorkerControl(tabA);

    await tabA.getByRole("combobox", { name: "Permissions" }).click();
    await tabA.getByRole("option", { name: "Full access" }).click();

    await expect(tabB.getByRole("combobox", { name: "Permissions" }))
      .toHaveText("Full access", { timeout: 10_000 });

    await expect(tabA.getByRole("combobox", { name: "Permissions" }))
      .toHaveText("Full access", { timeout: 5_000 });

    await tabA.close();
    await tabB.close();
  });
});
