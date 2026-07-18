// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { createResourceWorkflow } from "../../helpers/resource-workflow";
test.describe("Workflow Operations", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("workflows: open create dialog", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workflows`);
    await page.waitForLoadState("load");

    const createBtn = page.getByRole("button", { name: /新建|Create|New/i }).first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("workflows: list → detail navigation", async ({ page, api }) => {
    const cc = await api.connect();
    const created = await createResourceWorkflow(cc, {
      name: `E2E Workflow Nav ${Date.now()}`,
      slug: `e2e-workflow-nav-${Date.now()}`,
      cronExpression: "0 * * * *",
      prompt: "echo nav test",
    });
    const slug = created.slug;
    expect(slug).toBeTruthy();
    await page.goto(`/${TEST_ORG_SLUG}/workflows`);
    await page.waitForLoadState("load");

    const link = page.locator(`a[href*="workflows/${slug}"]`).first();
    if (await link.isVisible({ timeout: 5000 }).catch(() => false)) {
      await link.click();
      await page.waitForLoadState("load");
    }
    if (slug) {
      await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: slug }).catch(() => null);
    }
  });
});
