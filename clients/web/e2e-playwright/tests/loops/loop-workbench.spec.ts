import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Loop workbench", () => {
  test("keeps Blockly and LoopScript synchronized through invalid edits", async ({
    page,
  }) => {
    await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);

    await expect(
      page.getByRole("heading", { name: "Loop 工作台" }),
    ).toBeVisible();
    const runButton = page.getByRole("button", { name: "发给 AI 运行" });
    await expect(runButton).toBeEnabled();

    await page
      .getByRole("button", {
        name: "Edit text: 修复结算页税额计算，并补充边界测试。",
      })
      .click();
    const blocklyInput = page.locator(".blocklyHtmlInput");
    await blocklyInput.fill("浏览器积木联动验证");

    const codeEditor = page.locator(".cm-content");
    await expect(codeEditor).toContainText("浏览器积木联动验证");

    await page.getByRole("tab", { name: "代码" }).click();
    await expect(codeEditor).toBeEditable();
    const validSource = await codeEditor.innerText();
    const finalBrace = validSource.lastIndexOf("}");
    expect(finalBrace).toBeGreaterThan(0);
    const invalidSource =
      validSource.slice(0, finalBrace) + validSource.slice(finalBrace + 1);
    await codeEditor.fill(invalidSource);

    await expect(runButton).toBeDisabled();
    await expect(
      page.getByText("loop.syntax.unexpected-token"),
    ).toBeVisible();

    await codeEditor.fill(validSource);
    await expect(runButton).toBeEnabled();

    await codeEditor.fill(
      validSource.replace(
        "浏览器积木联动验证",
        "浏览器代码联动验证",
      ),
    );
    await expect(
      page.getByRole("button", {
        name: "Edit text: 浏览器代码联动验证",
      }),
    ).toBeVisible();
  });

  test("creates and starts a real GoalLoop from LoopScript", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);

    const runButton = page.getByRole("button", { name: "发给 AI 运行" });
    await expect(runButton).toBeEnabled();
    await runButton.click();

    await expect(page.getByText("尚未发起真实 GoalLoop。")).toBeHidden();
    await expect(page.getByText(/状态 active/)).toBeVisible();
    await expect(page.getByText(/Pod \S+/)).toBeVisible();
  });
});
