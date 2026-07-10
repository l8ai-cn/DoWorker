import { test, expect } from "@playwright/test";
import { LoginPage } from "../../pages/login.page";
import { TEST_USER } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Login Flow", () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    clearAuthRateLimit();
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test("login page displays all required elements", async () => {
    await expect(loginPage.usernameInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.submitButton).toBeVisible();
    await expect(loginPage.registerLink).toBeVisible();
  });

  test("successful login redirects to workspace", async ({ page }) => {
    await loginPage.login(TEST_USER.username, TEST_USER.password);

    await page.waitForURL((url) => !url.pathname.includes("/login"), {
      timeout: 30_000,
    });

    expect(page.url()).toMatch(/\/(dev-org|workspace|dashboard)/);
  });

  test("invalid credentials show error message", async ({ page }) => {
    await loginPage.login("wronguser", "wrongpassword");

    await page.waitForTimeout(2_000);
    expect(page.url()).toContain("/login");

    const error = await loginPage.getErrorText();
    expect(error).toBeTruthy();
  });

  test("empty form shows validation errors", async ({ page }) => {
    await loginPage.submitButton.click();

    expect(page.url()).toContain("/login");

    const usernameRequired = await loginPage.usernameInput.getAttribute("required");
    expect(usernameRequired).not.toBeNull();
  });
});
