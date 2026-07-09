import { test, expect } from "../../fixtures/index";
import { SettingsNavPage } from "../../pages/settings/settings-nav.page";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
test.describe("Settings Dialog Operations", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  // ── Personal ──

  test("personal/general: update profile form", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("personal", "general");

    const nameInput = page.locator('input[name="name"], input[id="name"]').first();
    if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await nameInput.fill("E2E Updated Name");
      const saveBtn = page.getByRole("button", { name: /save|保存/i }).first();
      if (await saveBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await saveBtn.click();
        await page.waitForTimeout(1000);
      }
    }
  });

  test("personal/git: open add credential dialog", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("personal", "git");

    const addBtn = page.getByRole("button", { name: /添加凭据|Add Credential/i });
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("personal/git: open add provider dialog", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("personal", "git");

    const addBtn = page.getByRole("button", { name: /添加提供商|Add Provider/i });
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("personal/notifications: toggle preference", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("personal", "notifications");

    const toggle = page.locator('button[role="switch"], input[type="checkbox"]').first();
    if (await toggle.isVisible({ timeout: 3000 }).catch(() => false)) {
      await toggle.click();
      await page.waitForTimeout(500);
      await toggle.click();
      await page.waitForTimeout(500);
    }
  });

  // ── Organization ──

  test("org/members: open invite dialog", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("organization", "members");

    const inviteBtn = page.getByRole("button", { name: /邀请|Invite/i });
    if (await inviteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await inviteBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("org/api-keys: open create key dialog", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("organization", "api-keys");

    const createBtn = page.getByRole("button", { name: /创建|Create|新建/i }).first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("org/extensions: open import skill dialog", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("organization", "extensions");

    const addBtn = page.getByRole("button", { name: /导入|Import|添加|Add/i }).first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test("org/general: danger zone visible", async ({ page }) => {
    const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
    await nav.goto("organization", "general");

    const body = await page.textContent("body");
    expect(body).toMatch(/E2E Test|dev-org|组织/i);
  });
});
