import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for the Workspace page.
 * URL: /{orgSlug}/workspace
 * Based on: web/src/app/(dashboard)/[org]/workspace/page.tsx
 */
export class WorkspacePage {
  readonly createPodButton: Locator;
  readonly emptyState: Locator;

  constructor(
    private page: Page,
    private orgSlug: string,
  ) {
    this.createPodButton = page.getByTestId("workspace-create-pod");
    this.emptyState = page.getByTestId("workspace-empty-state");
  }

  async goto(): Promise<void> {
    await this.page.goto(`/${this.orgSlug}/workspace`);
    await this.page.waitForLoadState("load");
  }

  /** Check if the terminal grid area exists. */
  async hasTerminalGrid(): Promise<boolean> {
    const grid = this.page.locator(
      '[data-testid="terminal-grid"], .xterm, [role="terminal"], .terminal-container'
    ).first();
    return grid.isVisible().catch(() => false);
  }

  /** Check if any pod tabs are visible. */
  async getPodTabCount(): Promise<number> {
    return this.page
      .locator('[data-testid="pod-tab"], button[role="tab"]')
      .count();
  }

  /** Check if empty state is visible. */
  async isEmptyState(): Promise<boolean> {
    return this.emptyState.isVisible().catch(() => false);
  }

  /** Open the create pod modal. */
  async openCreatePodModal(): Promise<void> {
    await this.createPodButton.click();
    await this.page
      .locator('[role="dialog"]')
      .first()
      .waitFor({ state: "visible" });
  }
}
