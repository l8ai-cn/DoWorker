import { expect, type Page } from "@playwright/test";

export class CreateWorkerPage {
  constructor(
    private page: Page,
    private orgSlug: string,
  ) {}

  async goto(initialWorkerTypeSlug?: string): Promise<void> {
    const query = initialWorkerTypeSlug
      ? `?image=${encodeURIComponent(initialWorkerTypeSlug)}`
      : "";
    await this.page.goto(`/${this.orgSlug}/workers/new${query}`);
    await this.page.waitForLoadState("domcontentloaded");
    await this.page
      .getByRole("heading", { name: /create worker|创建 worker/i })
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async selectRuntimeBundles(bundleNames: string[]): Promise<void> {
    await this.openWorkspaceStep();
    const section = this.runtimeBundleSection();
    const checkboxes = section.locator('input[type="checkbox"]');
    for (let index = 0; index < await checkboxes.count(); index += 1) {
      const checkbox = checkboxes.nth(index);
      if (await checkbox.isChecked()) await checkbox.uncheck();
    }
    for (const name of bundleNames) {
      const checkbox = section
        .locator("label", { hasText: name })
        .locator('input[type="checkbox"]')
        .first();
      await expect(checkbox).toBeVisible({ timeout: 15_000 });
      if (!(await checkbox.isChecked())) await checkbox.check();
    }
  }

  async selectPtyMode(): Promise<void> {
    await this.openWorkspaceStep();
    const interactive = this.page.getByTestId("automation-level-interactive");
    await interactive.waitFor({ state: "visible", timeout: 15_000 });
    await interactive.click();
    const pty = this.page.getByRole("button", {
      name: /^(Terminal \(PTY\)|终端 \(PTY\))$/i,
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
    await this.openPreflightStep();
    const button = this.submitButton();
    await expect(button).toBeEnabled({ timeout: 15_000 });
    await button.click();
  }

  async waitForWorkspace(timeoutMs = 15_000): Promise<void> {
    await this.page.waitForURL(
      new RegExp(`/${this.orgSlug}/workspace\\?pod=`),
      { timeout: timeoutMs },
    );
  }

  private async openWorkspaceStep(): Promise<void> {
    const task = this.page.locator("#worker-initial-task");
    if (await task.isVisible()) return;
    await this.next();
    await this.next();
    await expect(task).toBeVisible({ timeout: 15_000 });
  }

  private async openPreflightStep(): Promise<void> {
    await this.openWorkspaceStep();
    await this.next();
  }

  private async next(): Promise<void> {
    const next = this.page.getByRole("button", { name: /^(Next|下一步)$/i });
    await expect(next).toBeEnabled({ timeout: 15_000 });
    await next.click();
  }

  private runtimeBundleSection() {
    return this.page
      .locator("label", {
        hasText: /^(Select runtime bundles|选择运行时包)$/,
      })
      .first()
      .locator("..");
  }
}
