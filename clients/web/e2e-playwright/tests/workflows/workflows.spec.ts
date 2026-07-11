// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Workflows API & UI", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("list workflows", async ({ api }) => {
    const cc = await api.connect();
    const { items } = await cc.workflow.listWorkflows({ orgSlug: TEST_ORG_SLUG }) as { items: unknown[] };
    expect(Array.isArray(items)).toBe(true);
  });

  test("workflows page loads in UI", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workflows`);
    await page.waitForLoadState("load");
    const body = await page.textContent("body");
    expect(body).toMatch(/workflow|循环|定时/i);
  });
});
