// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { SettingsNavPage } from "../../pages/settings/settings-nav.page";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Organization Extensions Settings", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("API: list authored skills catalog", async ({ api }) => {
    const res = await api.get(`/api/v1/orgs/${TEST_ORG_SLUG}/authored-skills`);
    expect(res.ok).toBe(true);
    const { skills } = await res.json() as { skills: unknown[] };
    expect(Array.isArray(skills)).toBe(true);
  });

  test("UI: extensions settings page loads without errors", async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text());
    });

    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("organization", "extensions");

    const body = await page.textContent("body");
    expect(body).toMatch(/extension|skill|扩展|技能/i);

    const jsonErrors = consoleErrors.filter(
      (e) => e.includes("missing field") || e.includes("is not valid JSON")
    );
    expect(jsonErrors).toHaveLength(0);
  });
});
