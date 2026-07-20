import { expect, test } from "../../fixtures/index";
import { getWebBaseUrl, TEST_ORG_SLUG } from "../../helpers/env";
import {
  cleanupLoopRuntimeFixture,
  createLoopRuntimeFixture,
} from "../../helpers/loop-runtime-fixture";
import { LOCALE_COOKIE } from "../../../src/lib/i18n/config";
import {
  collectFailedRequests,
  doubleClickBlocklyBackground,
  openLocalizedLoopWorkbench,
  overflowingText,
  resetLoopSource,
} from "./loop-workbench-browser-helpers";

const ZH_LOOP_LABELS = {
  blocks: "积木",
  code: "代码",
  run: "运行循环",
  valid: "有效",
} as const;

const LOCALE_CASES = [
  {
    locale: "zh",
    title: "循环工作台",
    blocks: "积木",
    code: "代码",
    run: "运行循环",
    valid: "有效",
    custom: "创建自定义积木",
    failure: "失败处理",
    ai: "AI 助手",
  },
  {
    locale: "en",
    title: "Loop workbench",
    blocks: "Blocks",
    code: "Code",
    run: "Run loop",
    valid: "Valid",
    custom: "Create custom block",
    failure: "Failure handling",
    ai: "AI assistant",
  },
] as const;

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

    await resetLoopSource(page, ZH_LOOP_LABELS);
    const runButton = page.getByRole("button", { name: "运行循环" });

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

  test("creates custom blocks without putting Worker into the Loop AST", async ({
    page,
  }) => {
    await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);
    await expect(
      page.getByRole("heading", { name: "循环工作台" }),
    ).toBeVisible();
    await resetLoopSource(page, ZH_LOOP_LABELS);

    await doubleClickBlocklyBackground(page);
    await page.getByRole("button", { name: "创建自定义积木" }).click();
    const slug = `ppt-e2e-${Date.now()}`;
    const label = `专业 PPT ${slug.slice(-6)}`;
    const dialog = page.getByRole("dialog", { name: "创建自定义积木" });
    await dialog.getByLabel("积木名称").fill(label);
    await dialog.getByLabel("积木标识").fill(slug);
    await dialog.getByLabel("任务模板").fill("制作 {{topic}} 的专业 PPT");
    await dialog.getByLabel("验证命令模板").fill("test -f {{file}}");
    await dialog.getByLabel("验收说明模板").fill("{{file}} 存在且可打开");
    const created = page.waitForResponse((response) =>
      response.url().includes("BlockstoreService/ApplyOps") &&
      response.request().method() === "POST" &&
      response.ok(),
    );
    await dialog.getByRole("button", { name: "创建积木" }).click();
    await created;

    await expect(page.getByText("有效").first()).toBeVisible();
    expect(page.url()).not.toContain("loopCustomBlocks=");
    const codeEditor = page.locator(".cm-content");
    const workspace = page.getByLabel("Blockly Workspace");
    await expect(codeEditor).toContainText(`agent ${slug}-task`);
    await expect(workspace.getByText(label, { exact: true })).toBeVisible();
    const toolbox = page.locator(".blocklyToolbox");
    await toolbox.locator(".blocklyToolboxCategory", {
      hasText: "我的积木",
    }).click();
    await expect(
      page.locator(".blocklyToolboxFlyout .blocklyDraggable"),
    ).toBeVisible();

    await expect(page.getByText("有效").first()).toBeVisible();
    await expect(workspace.getByText("topic", { exact: true })).toBeVisible();
    await expect(workspace.getByText("file", { exact: true })).toBeVisible();
    await workspace.getByText("topic", { exact: true }).click();
    await page.locator(".blocklyHtmlInput").fill("季度复盘");
    await workspace.getByText("file", { exact: true }).click();
    await page.locator(".blocklyHtmlInput").fill("output.pptx");

    await expect(codeEditor).toContainText("制作 季度复盘 的专业 PPT");
    await expect(codeEditor).toContainText("test -f output.pptx");
    const source = await codeEditor.innerText();
    expect(source.toLowerCase()).not.toContain("worker");
    expect(source).not.toContain("invalid-block-structure");
    expect(source).toContain(`agent ${slug}-task`);
    expect(source).toContain(`verify ${slug}-check`);

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(
      page.getByRole("heading", { name: "循环工作台" }),
    ).toBeVisible();
    await resetLoopSource(page, ZH_LOOP_LABELS);
    await doubleClickBlocklyBackground(page);
    await expect(page.getByRole("button", { name: label })).toBeVisible();
    await page.getByRole("button", { name: label }).click();
    await expect(codeEditor).toContainText(`agent ${slug}-task`);
  });

  test("keeps Loop projection equivalent across Chinese and English workbenches", async ({
    page,
    context,
  }) => {
    const failedRequests = collectFailedRequests(page);
    const sources: string[] = [];

    for (const localeCase of LOCALE_CASES) {
      const { title, blocks, code, run, valid, custom, failure, ai } = localeCase;
      await openLocalizedLoopWorkbench(page, context, localeCase);
      await expect(page.getByRole("heading", { name: title })).toBeVisible();
      await resetLoopSource(page, { blocks, code, run, valid });
      const codeEditor = page.locator(".cm-content");
      sources.push(await codeEditor.innerText());

      await page.getByRole("button", { name: ai }).click();
      await expect(page.getByRole("dialog")).toBeVisible();
      expect(await codeEditor.innerText()).toBe(sources.at(-1));
      await page.keyboard.press("Escape");

      await doubleClickBlocklyBackground(page);
      await expect(page.getByRole("button", { name: custom })).toBeVisible();
      expect(await overflowingText(page, [custom, failure])).toEqual([]);
      await page.getByRole("button", { name: failure }).click();
      await expect(codeEditor).toContainText("invalid-block-structure");
      expect((await codeEditor.innerText()).toLowerCase()).not.toContain("worker");
    }

    expect(new Set(sources).size).toBe(1);
    expect(failedRequests()).toEqual([]);
  });

  test("keeps Chinese and English Loop workbenches usable on mobile", async ({
    page,
    context,
  }) => {
    const failedRequests = collectFailedRequests(page);
    await page.setViewportSize({ width: 390, height: 844 });

    for (const localeCase of LOCALE_CASES) {
      const { title, blocks, code, run, valid, custom } = localeCase;
      await openLocalizedLoopWorkbench(page, context, localeCase);
      await expect(page.getByRole("heading", { name: title })).toBeVisible();
      await resetLoopSource(page, { blocks, code, run, valid });
      await doubleClickBlocklyBackground(page);
      await expect(page.getByRole("button", { name: custom })).toBeVisible();
      expect(await overflowingText(page, [title, blocks, code, run, custom]))
        .toEqual([]);
      expect((await page.locator(".cm-content").innerText()).toLowerCase())
        .not.toContain("worker");
    }
    expect(failedRequests()).toEqual([]);
  });

  test("applies a GoalLoop resource before starting execution", async ({
    page,
    db,
  }) => {
    const runtime = createLoopRuntimeFixture(db);
    try {
      await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);
      await resetLoopSource(page, ZH_LOOP_LABELS);
      await page.getByRole("tab", { name: "代码" }).click();
      const codeEditor = page.locator(".cm-content");
      await expect(codeEditor).toBeEditable();
      await codeEditor.fill(
        (await codeEditor.innerText()).replaceAll(
          "checkout-fix",
          runtime.goalLoopName,
        ),
      );

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
        page.getByText("循环执行现在必须先通过资源编排 validate-plan-apply，再启动。"),
      ).toHaveCount(0);
      await expect.poll(() => db.queryValue(`
        SELECT count(*)
        FROM goal_loops
        WHERE worker_spec_snapshot_id = ${runtime.snapshotId}
          AND orchestration_resource_id IS NOT NULL
          AND orchestration_resource_revision = 1
      `)).toBe("1");
      expect(db.queryValue(`
        SELECT count(*)
        FROM orchestration_resources
        WHERE kind = 'GoalLoop'
          AND name = '${runtime.goalLoopName}'
          AND namespace = '${TEST_ORG_SLUG}'
      `)).toBe("1");
    } finally {
      cleanupLoopRuntimeFixture(db, runtime);
    }
  });
});
