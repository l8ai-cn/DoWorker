import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for the Login page.
 * Selectors based on: web/src/app/(auth)/login/page.tsx
 */
export class LoginPage {
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly registerLink: Locator;
  readonly errorMessage: Locator;

  constructor(private page: Page) {
    this.usernameInput = page.locator("#username");
    this.passwordInput = page.locator("#password");
    this.submitButton = page.locator('button[type="submit"]');
    this.registerLink = page.getByRole("link", { name: /sign up|register/i });
    this.errorMessage = page.locator(".text-destructive");
  }

  async goto(): Promise<void> {
    await this.page.goto("/login");
    await this.usernameInput.waitFor({ state: "visible" });
  }

  async login(username: string, password: string): Promise<void> {
    await this.page.waitForTimeout(250);
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }

  async getErrorText(): Promise<string | null> {
    if (await this.errorMessage.isVisible()) {
      return this.errorMessage.textContent();
    }
    return null;
  }
}
