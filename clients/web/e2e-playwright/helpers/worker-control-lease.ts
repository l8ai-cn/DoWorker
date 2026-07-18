import { expect, type Page } from "@playwright/test";

export async function takeWorkerControl(page: Page): Promise<void> {
  const button = page.getByRole("button", { name: /^take control$/i });
  await expect(button).toBeVisible({ timeout: 30_000 });
  await button.click();
  await expect(button).toBeHidden({ timeout: 30_000 });
}
