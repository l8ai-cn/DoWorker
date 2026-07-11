import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

/**
 * Workflow create/edit dialog × EnvBundle UI flow.
 *
 * Complements the API-level coverage in workflow-env-bundle.spec.ts:
 *
 *   - This file drives the actual WorkflowCreateDialog, asserting runtime bundle
 *     picks survive the round-trip to the backend and edit mode reconciles
 *     `used_env_bundles` back into checked rows.
 *   - workflow-env-bundle.spec.ts asserts the wire contract (POST/PUT round-trip,
 *     `[]` clears, dangling names tolerated).
 */
const BUNDLE_PREFIX = "E2E WorkflowUI Bundle";
const LOOP_PREFIX = "E2E WorkflowUI Workflow";
const LOOP_AGENT_SLUG = "e2e-echo";

function unique(prefix: string, label: string): string {
  return `${prefix} ${label} ${Date.now()}`;
}

async function openCreateWorkflowDialog(page: import("@playwright/test").Page) {
  await page.goto(`/${TEST_ORG_SLUG}/workflows`);
  // Use "load" not "networkidle" — DashboardShell holds an open
  // RealtimeProvider WebSocket, so the page never reaches networkidle.
  await page.waitForLoadState("load");
  const btn = page
    .getByRole("button", { name: /create workflow|新建 ?workflow|创建 ?workflow|创建你的第一个/i })
    .first();
  await btn.waitFor({ state: "visible", timeout: 15_000 });
  await btn.click();
  await page.locator('[data-dialog-overlay]').first().waitFor({ state: "visible" });
}

async function expandAdvancedOptions(page: import("@playwright/test").Page) {
  const adv = page
    .locator('[data-dialog-overlay]')
    .getByRole("button", { name: /advanced options|高级选项/i });
  if (await adv.isVisible().catch(() => false)) {
    const state = await adv.getAttribute("data-state");
    if (state !== "open") await adv.click();
  }
}

async function pickDialogSelectOption(
  page: import("@playwright/test").Page,
  triggerId: string,
  optionValue: string,
) {
  const overlay = page.locator("[data-dialog-overlay]").first();
  const trigger = overlay.locator(`#${triggerId}`).first();
  await trigger.waitFor({ state: "visible", timeout: 15_000 });
  await trigger.click();
  await overlay
    .locator(`[role="option"][data-option-value="${optionValue}"]`)
    .first()
    .click();
}

test.describe("Workflow dialog — EnvBundle binding UI", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  test("create flow: runtime checkbox binds and submits selected bundle", async ({
    page,
    api,
    db,
  }) => {
    const runtimeName = unique(BUNDLE_PREFIX, "runtime");
    const workflowName = unique(LOOP_PREFIX, "create");

    db.cleanup(`DELETE FROM env_bundles WHERE name LIKE '${BUNDLE_PREFIX}%'`);

    const cc = await api.connect();
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: LOOP_AGENT_SLUG,
      name: runtimeName,
      kind: "runtime",
      data: { CLAUDE_LOG_LEVEL: "debug" },
    }) as { id: bigint };
    const runtimeId = runtime.id;

    let workflowSlug: string | undefined;
    try {
      await openCreateWorkflowDialog(page);

      await page
        .locator('[data-dialog-overlay] input[placeholder="daily-code-review"]')
        .first()
        .fill(workflowName);

      await pickDialogSelectOption(page, "worker-image-select", LOOP_AGENT_SLUG);

      const promptInput = page
        .locator('[data-dialog-overlay] textarea#prompt-input')
        .first();
      await promptInput.waitFor({ state: "visible", timeout: 5000 });
      await promptInput.fill("run nightly tests");

      await expandAdvancedOptions(page);

      const overlay = page.locator('[data-dialog-overlay]');

      await expect(overlay.locator('select#credential-bundle-select')).toHaveCount(0);

      // Runtime picker is a checkbox list. Toggle the seeded runtime bundle.
      const runtimeCheckbox = overlay
        .getByRole("checkbox", { name: new RegExp(runtimeName) })
        .first();
      await runtimeCheckbox.waitFor({ state: "visible", timeout: 5000 });
      if (!(await runtimeCheckbox.isChecked())) await runtimeCheckbox.click();

      await overlay
        .getByRole("button", { name: /^(Create Workflow|创建 ?Workflow)$/i })
        .click();

      // Backend should persist only the selected runtime bundle.
      await expect
        .poll(
          () => {
            const raw = db.queryValue(
              `SELECT array_to_string(used_env_bundles, ',') FROM workflows WHERE name = '${workflowName.replace(/'/g, "''")}'`
            );
            return raw ?? "";
          },
          { timeout: 10_000 }
        )
        .toBe(runtimeName);

      workflowSlug = db.queryValue(
        `SELECT slug FROM workflows WHERE name = '${workflowName.replace(/'/g, "''")}'`
      ) ?? undefined;
    } finally {
      if (workflowSlug) {
        await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug }).catch(() => null);
      }
      if (runtimeId) await cc.envBundle.deleteEnvBundle({ id: runtimeId }).catch(() => null);
      db.cleanup(`DELETE FROM env_bundles WHERE name LIKE '${BUNDLE_PREFIX}%'`);
    }
  });

  test("edit flow: existing used_env_bundles reconcile to runtime checkbox", async ({
    page,
    api,
    db,
  }) => {
    const runtimeName = unique(BUNDLE_PREFIX, "edit-runtime");
    const workflowName = unique(LOOP_PREFIX, "edit");

    db.cleanup(`DELETE FROM env_bundles WHERE name LIKE '${BUNDLE_PREFIX}%'`);

    const cc = await api.connect();
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: LOOP_AGENT_SLUG,
      name: runtimeName,
      kind: "runtime",
      data: { CLAUDE_LOG_LEVEL: "debug" },
    }) as { id: bigint };
    const runtimeId = runtime.id;

    const workflowRes = await cc.workflow.createWorkflow({
      orgSlug: TEST_ORG_SLUG,
      name: workflowName,
      agentSlug: LOOP_AGENT_SLUG,
      promptTemplate: "echo bound",
      usedEnvBundles: [runtimeName],
    }) as { slug: string };
    const workflowSlug = workflowRes.slug;

    try {
      await page.goto(`/${TEST_ORG_SLUG}/workflows/${workflowSlug}`);
      await page.waitForLoadState("load");

      await page
        .getByRole("heading", { name: workflowName, level: 1 })
        .waitFor({ state: "visible", timeout: 10_000 })
        .catch(() => {});

      const editBtn = page
        .getByRole("button")
        .filter({ hasText: /^(Edit|编辑)$/i })
        .first();
      await expect(editBtn, "workflow detail page must render an Edit button for the workflow creator")
        .toBeVisible({ timeout: 10_000 });
      await editBtn.click();
      await page.locator('[data-dialog-overlay]').first().waitFor({ state: "visible" });

      await expandAdvancedOptions(page);

      const overlay = page.locator('[data-dialog-overlay]');

      await expect(overlay.locator('select#credential-bundle-select')).toHaveCount(0);

      // Runtime checkbox should be pre-checked for the saved runtime bundle.
      const runtimeCheckbox = overlay
        .getByRole("checkbox", { name: new RegExp(runtimeName) })
        .first();
      await runtimeCheckbox.waitFor({ state: "visible", timeout: 5000 });
      expect(await runtimeCheckbox.isChecked()).toBe(true);
    } finally {
      if (workflowSlug) {
        await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug }).catch(() => null);
      }
      if (runtimeId) await cc.envBundle.deleteEnvBundle({ id: runtimeId }).catch(() => null);
      db.cleanup(`DELETE FROM env_bundles WHERE name LIKE '${BUNDLE_PREFIX}%'`);
    }
  });
});
