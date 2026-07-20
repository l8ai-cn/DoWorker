import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { setupAcpScenarioPage, takeWorkerControl } from "../../helpers/acp-spec-setup";

// Defensive-path coverage: every scenario here exercises an unhappy
// runner/agent boundary that should NOT crash the web UI or wedge the
// activity stream.
// See acp-ui-echo.spec.ts header — same r6 fix applies.
test.describe("ACP UI: error and degradation paths", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });
  test.afterEach(async () => { await terminateRegisteredE2EPods(); });

  test("tool_call_failed renders the failed status without crashing UI", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "tool_call_failed", prompt: "edit me",
    });

    await expect(page.getByText("Trying to edit: edit me")).toBeVisible({ timeout: 15_000 });
    const fileChange = page.locator("details").filter({ hasText: "Changed 1 file" }).first();
    await expect(fileChange).toBeVisible({ timeout: 15_000 });
    await takeWorkerControl(page);
    await fileChange.locator(":scope > summary").click();
    await expect(fileChange.getByText("file not found", { exact: true })).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("malformed_json output does not break subsequent valid messages", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "malformed_json", prompt: "garbled",
    });

    await expect(page.getByText("recovered: garbled")).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("log_warnings surfaces warn/error stderr lines in activity stream", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "log_warnings", prompt: "noisy run",
    });

    await expect(page.getByText(/degraded connection/i)).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Completed with warnings: noisy run")).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });

  test("fail_after_1s does not leave the UI wedged in a processing state", async ({ page, api, monitor }) => {
    const ctx = await setupAcpScenarioPage(page, api, monitor, {
      mode: "acp", scenario: "fail_after_1s",
    });

    await takeWorkerControl(page);
    const composer = page.getByLabel("Message the agent");
    await expect(composer).toBeEnabled({ timeout: 15_000 });
    await composer.fill("crash test");
    await composer.press("Enter");

    await expect(page.getByText("Will crash soon: crash test")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("heading", { name: "Spawn your first Pod" })).toBeVisible({ timeout: 15_000 });
    ctx.assertWasmHealthy();
  });
});
