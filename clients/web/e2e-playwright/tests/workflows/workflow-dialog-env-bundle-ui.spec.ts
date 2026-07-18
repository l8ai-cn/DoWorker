import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Worker template dependency editor", () => {
  test("keeps Skills, knowledge, and environment bundle references in one workspace section", async ({
    page,
  }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    await page.getByTestId("pill-tab-template").click();

    const editor = page.getByTestId("resource-editor");
    await expect(editor.getByText(/Skills|Skill|技能/i)).toBeVisible();
    await expect(editor.getByText(/Knowledge mounts|知识库挂载/i)).toBeVisible();
    await expect(editor.getByText(/Environment bundles|环境变量包/i)).toBeVisible();
  });

  test("exposes explicit resource sizing controls instead of an opaque profile only", async ({
    page,
  }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    await page.getByTestId("pill-tab-template").click();

    const editor = page.getByTestId("resource-editor");
    const resourceMode = editor.locator("button").filter({
      hasText: /^(profile|资源规格)$/i,
    });
    await expect(resourceMode).toBeVisible();
    await resourceMode.click();
    await expect(editor.getByRole("option", {
      name: /Custom limits|自定义限制/i,
    })).toBeVisible();
  });
});
