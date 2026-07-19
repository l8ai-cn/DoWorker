import path from "node:path";
import { chromium } from "@playwright/test";

export async function runBrowserPreflight(config, result) {
  let browser;
  try {
    browser = await chromium.launch({
      headless: true,
      executablePath: config.chromiumExecutable,
    });
  } catch (error) {
    throw new Error(`Chromium runtime is unavailable: ${error.message}`);
  }
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  try {
    await page.goto(`${config.webUrl}/login`, { waitUntil: "domcontentloaded" });
    await page.locator("#login-username").waitFor({ state: "visible" });
    await page.screenshot({
      path: path.join(config.evidenceDir, "01-login.png"),
      fullPage: true,
      caret: "initial",
    });
    await page.locator("#login-username").fill(config.username);
    await page.locator("#login-password").fill(config.password);
    await page.locator('button[type="submit"]').click();
    await page.waitForURL((url) => url.pathname !== "/login", { timeout: 20000 });
    page.on("console", (message) => {
      if (message.type() === "error") {
        result.failures.push(`browser console error: ${message.text()}`);
      }
    });
    page.on("pageerror", (error) => {
      result.failures.push(`browser page error: ${error.message}`);
    });
    page.on("response", (response) => {
      if (response.status() < 400) return;
      const type = response.request().resourceType();
      if (type === "document" || type === "fetch" || type === "xhr") {
        result.failures.push(`browser HTTP ${response.status()}: ${response.url()}`);
      }
    });
    await page.goto(
      `${config.webUrl}/${config.orgSlug}/settings?scope=organization&tab=extensions`,
      { waitUntil: "domcontentloaded" },
    );
    const importButton = page.getByRole("button", {
      name: /Import from Git|从 Git 导入|Import/i,
    });
    await importButton.waitFor({ state: "visible" });
    await importButton.click();
    await page.getByRole("button", { name: "Pattern Designer" }).waitFor({ state: "visible" });
    await page.screenshot({
      path: path.join(config.evidenceDir, "03-import-dialog.png"),
      fullPage: true,
      caret: "initial",
    });
    result.browser.status = "pattern_designer_import_option_visible";
  } finally {
    await browser.close();
  }
}
