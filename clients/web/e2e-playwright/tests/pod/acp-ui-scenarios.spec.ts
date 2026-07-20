import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { setupAcpScenarioPage, takeWorkerControl } from "../../helpers/acp-spec-setup";

// Scenario coverage for the universal mock agent through the shared
// AgentWorkspace activity and approval surfaces.
test.describe("ACP UI: mock agent scenario matrix", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("streaming_3 emits three chunks concatenated in the activity stream", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "streaming_3", prompt: "hello",
    });

    await expect(page.getByText(/streaming: hello\s+\(done\)/)).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("thinking_then_answer renders reasoning evidence and final content", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "thinking_then_answer", prompt: "what is 2+2",
    });

    await expect(page.getByText(/Let me analyze the prompt: what is 2\+2/)).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Answer to: what is 2+2")).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("tool_call_edit renders a completed file-change activity", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "tool_call_edit", prompt: "edit me",
    });

    await expect(page.getByText("Changed 1 file", { exact: true })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Editing file for: edit me")).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("permission_request_edit shows permission dialog and approval completes the tool", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "permission_request_edit", prompt: "edit me carefully",
    });

    await expect(page.getByText("Edit", { exact: true }).first()).toBeVisible({ timeout: 15_000 });
    await takeWorkerControl(page);
    await page.getByRole("button", { name: /Approve/i }).first().click();
    const fileChange = page.locator("details").filter({ hasText: "Changed 1 file" }).first();
    await expect(fileChange).toBeVisible({ timeout: 10_000 });
    await fileChange.locator(":scope > summary").click();
    await fileChange.getByText("Details", { exact: true }).click();
    await expect(fileChange.getByText("Edited 1 line", { exact: true })).toBeVisible({ timeout: 10_000 });
    ctx.assertWasmHealthy();
  });
});
