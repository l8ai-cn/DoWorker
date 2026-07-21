import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Worker resource editor", () => {
  test("shows the Worker invocation contract and blocks incomplete creation", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    await page.getByTestId("pill-tab-template").click();

    const editor = page.getByTestId("resource-editor");
    await expect(editor).toBeVisible();
    await expect(editor.getByLabel(/Resource name|资源名称/i)).toBeVisible();
    await expect(editor.getByRole("combobox", {
      name: /Model binding|模型绑定/i,
    })).toBeVisible();

    const createButton = editor.getByRole("button", {
      name: /^(Create Worker|创建 Worker)$/i,
    });
    await expect(createButton).toBeDisabled();

    await editor.getByLabel(/Resource name|资源名称/i).fill("e2e-worker-invocation");
    await expect(createButton).toBeDisabled();
  });
});
