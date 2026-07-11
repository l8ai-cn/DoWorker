import type { Page } from "@playwright/test";

export class CreateWorkerPage {
  constructor(
    private page: Page,
    private orgSlug: string,
  ) {}

  async goto(): Promise<void> {
    await this.page.goto(`/${this.orgSlug}/workers/new`);
    await this.page.waitForLoadState("domcontentloaded");
    await this.page
      .getByRole("heading", { name: /create worker|创建 worker/i })
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async selectImage(imageSlug: string): Promise<void> {
    const trigger = this.page.locator("#worker-image-select");
    await trigger.waitFor({ state: "visible", timeout: 15_000 });
    await trigger.click();
    await this.page
      .locator(`[role="option"][data-option-value="${imageSlug}"]`)
      .click();
  }

  async selectRuntimeBundles(bundleNames: string[]): Promise<void> {
    const checkboxes = this.page.locator(
      '[data-testid="worker-credential-model-section"] input[type="checkbox"]',
    );
    for (let i = 0; i < await checkboxes.count(); i += 1) {
      const checkbox = checkboxes.nth(i);
      if (await checkbox.isChecked()) await checkbox.click();
    }
    for (const name of bundleNames) {
      const row = this.page.locator("label").filter({ hasText: name }).first();
      await row.locator('input[type="checkbox"]').click();
    }
  }

  async selectPtyMode(): Promise<void> {
    const interactive = this.page.getByTestId("automation-level-interactive");
    await interactive.waitFor({ state: "visible", timeout: 15_000 });
    await interactive.click();
    const pty = this.page.getByRole("button", {
      name: /terminal \(pty\)|终端 \(pty\)/i,
    });
    await pty.waitFor({ state: "visible", timeout: 15_000 });
    await pty.click();
  }

  private submitButton() {
    return this.page
      .getByRole("button", { name: /^create worker$/i })
      .last();
  }

  async submit(): Promise<void> {
    const button = this.submitButton();
    await button.waitFor({ state: "visible", timeout: 15_000 });
    const handle = await button.elementHandle();
    if (!handle) throw new Error("Create Worker submit button is not mounted");
    await this.page.waitForFunction(
      (element) => !(element as HTMLButtonElement).disabled,
      handle,
      { timeout: 15_000 },
    );
    await button.click();
  }

  async waitForWorkspace(timeoutMs = 15_000): Promise<void> {
    await this.page.waitForURL(
      new RegExp(`/${this.orgSlug}/workspace\\?pod=`),
      { timeout: timeoutMs },
    );
  }
}
