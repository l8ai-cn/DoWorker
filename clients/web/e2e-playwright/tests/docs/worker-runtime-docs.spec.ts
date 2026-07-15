import { expect, test } from "../../fixtures";

const WORKER_NAMES = [
  "Aider",
  "Claude Code",
  "Codex CLI",
  "Cursor CLI",
  "Do Agent",
  "Gemini CLI",
  "Grok Build",
  "Hermes",
  "Loopal",
  "MiniMax CLI",
  "OpenClaw",
  "OpenCode",
];

test("public Worker docs use the runtime catalog instead of legacy terminology", async ({
  page,
}) => {
  await page.goto("/docs");

  await expect(page.locator('a[href="/docs/concepts/workers"]').first()).toBeVisible();
  await expect(page.locator('a[href="/docs/features/agentpod"]')).toHaveCount(0);
  await expect(page.locator('a[href="/docs/features/mesh"]')).toHaveCount(0);
  await expect(page.locator("main")).not.toContainText("AgentPod");
  await expect(page.locator("main")).not.toContainText("Aider");

  await page.goto("/docs/concepts/workers");

  const workerDocs = page.locator("main");
  for (const name of WORKER_NAMES) {
    await expect(workerDocs).toContainText(name);
  }
  await expect(workerDocs).not.toContainText("AgentPod");
  await expect(workerDocs).toContainText(
    "Local product flow verified; release image blocked",
  );
  await expect(workerDocs).toContainText(
    "Configured release digest cannot be pulled",
  );
  const geminiCard = page.locator("article.surface-card").filter({
    has: page.getByRole("heading", { name: "Gemini CLI", exact: true }),
  });
  await expect(geminiCard).toContainText("GEMINI_API_KEY");
  await expect(geminiCard).not.toContainText("GOOGLE_API_KEY");
});

test("legacy worker docs redirect and the catalog remains responsive", async ({ page }) => {
  await page.goto("/docs/features/agentpod");
  await expect(page).toHaveURL(/\/docs\/concepts\/workers$/);

  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/docs/concepts/workers");

  await expect(page.getByRole("banner")).toBeVisible();
  await expect(page.getByRole("button", { name: /menu|菜单/i })).toBeVisible();
  await expect
    .poll(() =>
      page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
  )
    .toBe(true);
});

test("docs header uses viewport padding on wide screens", async ({ page }) => {
  await page.setViewportSize({ width: 2048, height: 1200 });
  await page.goto("/docs/concepts/workers");

  const logo = page.getByRole("link", { name: "Do Worker" });
  const bounds = await logo.boundingBox();
  expect(bounds).not.toBeNull();
  expect(bounds?.x).toBeLessThanOrEqual(32);
});
