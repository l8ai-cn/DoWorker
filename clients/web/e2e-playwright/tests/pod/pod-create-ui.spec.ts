import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

/**
 * Public worker creation requires a configured model resource. Internal
 * e2e-echo pods use the typed test contract and are not public WorkerSpec
 * options, so the browser must surface this prerequisite rather than expose
 * the internal agent.
 */
test.describe("Create Worker UI", () => {
  test("shows the model-resource prerequisite and prevents progression", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    await expect(
      page.getByRole("heading", { name: /create worker|创建 worker/i }),
    ).toBeVisible();
    await expect(page.locator('[data-runtime-field="model"]')).toContainText(
      /no compatible model resources are available/i,
    );
    await expect(
      page.getByRole("button", { name: /^(Next|下一步)$/i }),
    ).toBeDisabled();
  });
});
