import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Worker template editor guards", () => {
  test("exposes runtime, placement, and workspace dependency fields", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    await page.getByTestId("pill-tab-template").click();

    const editor = page.getByTestId("resource-editor");
    await expect(editor).toContainText(/Worker template|Worker 模板/i);
    await expect(editor.getByLabel(/Worker type|Worker 类型/i)).toBeVisible();
    await expect(editor.getByLabel(/Deployment mode|部署模式|资源使用方式/i)).toBeVisible();
    await expect(editor.getByLabel(/Compute target|计算目标|算力目标/i).first()).toBeVisible();
    await expect(editor.getByText(/Resource allocation|Resource mode|资源分配|资源模式/i)).toBeVisible();
    await expect(editor.getByText(/Skills|技能/i)).toBeVisible();
    await expect(editor.getByText(/Environment bundles|环境变量包/i)).toBeVisible();
  });
});
