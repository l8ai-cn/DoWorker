import type { Locator, Page } from "@playwright/test";

/**
 * Create-pod form surface: legacy `[role=dialog]` modal or full-page
 * wizard at `/{org}/workers/new` (same CreatePodForm).
 */
export class CreatePodModal {
  constructor(private page: Page) {}

  /** Dialog when present; otherwise the wizard column (excludes NL create). */
  private root(): Locator {
    return this.page
      .locator('[role="dialog"]')
      .filter({ has: this.page.locator("#worker-image-select") })
      .or(
        this.page
          .locator(".flex-1.min-w-0")
          .filter({ has: this.page.locator("#worker-image-select") }),
      )
      .first();
  }

  async waitForOpen(): Promise<void> {
    await this.page
      .locator("#worker-image-select")
      .first()
      .waitFor({ state: "visible", timeout: 20_000 });
  }

  /**
   * Modal: dialog hides. Page: success navigates away from /workers/new.
   */
  async waitForClosed(timeoutMs = 15_000): Promise<void> {
    const dialog = this.page.locator('[role="dialog"]').filter({
      has: this.page.locator("#worker-image-select"),
    });
    if (await dialog.count()) {
      await dialog.first().waitFor({ state: "hidden", timeout: timeoutMs });
      return;
    }
    await this.page.waitForURL((url) => !url.pathname.includes("/workers/new"), {
      timeout: timeoutMs,
    });
  }

  private async pickSelectOption(triggerId: string, optionValue: string): Promise<void> {
    const root = this.root();
    await root.locator(`#${triggerId}`).click();
    // Options render inside the Select's relative wrapper (not a portal).
    await this.page
      .locator(`[role="option"][data-option-value="${optionValue}"]`)
      .first()
      .click();
  }

  async selectImage(imageSlug: string): Promise<void> {
    const trigger = this.root().locator("#worker-image-select");
    await trigger.waitFor({ state: "visible", timeout: 15_000 });
    await this.pickSelectOption("worker-image-select", imageSlug);
  }

  /** @deprecated use selectImage */
  async selectAgent(agentSlug: string): Promise<void> {
    return this.selectImage(agentSlug);
  }

  async fillPrompt(prompt: string): Promise<void> {
    await this.root().locator("textarea").first().fill(prompt);
  }

  /**
   * Legacy AdvancedOptions / current "More options" disclosure.
   * Credential + runtime pickers live on step 1 after image select —
   * this is only needed for alias / lifecycle fields.
   */
  async expandAdvancedOptions(): Promise<void> {
    const trigger = this.root().getByRole("button", {
      name: /advanced options|more options|高级选项|更多选项/i,
    });
    if (!(await trigger.isVisible().catch(() => false))) return;
    const state = await trigger.getAttribute("data-state");
    if (state !== "open") {
      await trigger.click();
    }
  }

  async selectCredential(bundleName: string): Promise<void> {
    await this.root()
      .locator("#credential-bundle-select")
      .waitFor({ state: "visible", timeout: 15_000 });
    await this.pickSelectOption("credential-bundle-select", bundleName);
  }

  async toggleRuntimeBundle(bundleName: string): Promise<void> {
    const row = this.root().locator("label", { hasText: bundleName }).first();
    await row.locator('input[type="checkbox"]').first().click();
  }

  async selectRuntimeBundles(bundleNames: string[]): Promise<void> {
    const section = this.root().locator(
      '[data-testid="worker-credential-model-section"]',
    );
    await section.waitFor({ state: "visible", timeout: 15_000 });
    const all = section.locator('input[type="checkbox"]');
    const count = await all.count();
    for (let i = 0; i < count; i += 1) {
      const box = all.nth(i);
      if (await box.isChecked()) await box.click();
    }
    for (const name of bundleNames) {
      await this.toggleRuntimeBundle(name);
    }
  }

  credentialTrigger(): Locator {
    return this.root().locator("#credential-bundle-select");
  }

  credentialOption(bundleName: string): Locator {
    return this.page.locator(
      `[role="option"][data-option-value="${bundleName}"]`,
    );
  }

  runtimeBundleLabel(bundleName: string): Locator {
    return this.root().locator("label", { hasText: bundleName });
  }

  async submit(): Promise<void> {
    await this.root()
      .getByRole("button", { name: /create worker|创建 worker|^create$|^创建$/i })
      .click();
  }

  async cancel(): Promise<void> {
    const btn = this.root().getByRole("button", { name: /cancel|取消/i });
    if (await btn.isVisible()) await btn.click();
  }
}
