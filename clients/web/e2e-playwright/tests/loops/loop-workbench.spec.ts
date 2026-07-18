import { expect, test } from "../../fixtures/index";
import { getWebBaseUrl, TEST_ORG_SLUG } from "../../helpers/env";
import {
  cleanupLoopRuntimeFixture,
  createLoopRuntimeFixture,
} from "../../helpers/loop-runtime-fixture";
import { LOCALE_COOKIE } from "../../../src/lib/i18n/config";

test.describe("Loop workbench", () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([
      {
        name: LOCALE_COOKIE,
        value: "zh",
        url: getWebBaseUrl(),
      },
    ]);
  });

  test("keeps Blockly and LoopScript synchronized through invalid edits", async ({
    page,
  }) => {
    const externalBlocklyMediaRequests: string[] = [];
    page.on("request", (request) => {
      if (request.url().includes("blockly-demo.appspot.com")) {
        externalBlocklyMediaRequests.push(request.url());
      }
    });

    await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);

    await expect(
      page.getByRole("heading", { name: "循环工作台" }),
    ).toBeVisible();
    const runButton = page.getByRole("button", { name: "运行循环" });
    await expect(runButton).toBeEnabled();

    await page
      .getByText("修复结算页税额计算，并补充边界测试。", { exact: true })
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
    await expect(page.getByText("循环脚本结构不符合语法")).toBeVisible();

    await codeEditor.fill(validSource);
    await expect(runButton).toBeEnabled();

    await codeEditor.fill(
      validSource.replace("浏览器积木联动验证", "浏览器代码联动验证"),
    );
    await expect(
      page.getByText("浏览器代码联动验证", { exact: true }),
    ).toBeVisible();
    expect(externalBlocklyMediaRequests).toEqual([]);
  });

  test("requires resource apply before starting a GoalLoop", async ({
    page,
    db,
    monitor,
  }) => {
    monitor.allow(/GoalLoopService\/RunLoopProgram/);
    const runtime = createLoopRuntimeFixture(db);
    try {
      await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);

      const runButton = page.getByRole("button", { name: "运行循环" });
      await expect(runButton).toBeEnabled();
      await runButton.click();

      const dialog = page.getByRole("dialog", { name: "选择运行环境" });
      await expect(dialog).toContainText(
        "运行环境只在本次启动时绑定，不属于循环编排。",
      );
      const startButton = dialog.getByRole("button", { name: "启动循环" });
      await expect(startButton).toBeDisabled();
      await dialog.getByRole("button", { name: "选择运行环境" }).click();
      await dialog
        .getByRole("option", { name: new RegExp(runtime.alias) })
        .click();
      await expect(startButton).toBeEnabled();
      await startButton.click();

      await expect(
        page.getByText("循环启动失败，请确认运行环境仍然可用"),
      ).toBeVisible();
      expect(db.queryValue(`
        SELECT count(*)
        FROM goal_loops
        WHERE worker_spec_snapshot_id = ${runtime.snapshotId}
      `)).toBe("0");
    } finally {
      cleanupLoopRuntimeFixture(db, runtime);
    }
  });
});
