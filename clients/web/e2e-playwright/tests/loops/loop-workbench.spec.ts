import type { BrowserContext, Page, Request } from "@playwright/test";
import { expect, test } from "../../fixtures/index";
import { getWebBaseUrl, TEST_ORG_SLUG } from "../../helpers/env";
import {
  cleanupLoopRuntimeFixture,
  createLoopRuntimeFixture,
} from "../../helpers/loop-runtime-fixture";
import { LOCALE_COOKIE } from "../../../src/lib/i18n/config";

const DEFAULT_LOOP_SOURCE = `@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
  }
  on_failure pause
}`;

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
  {
    locale: "de",
    title: "Loop-Workbench",
    blocks: "Blöcke",
    code: "Code",
    run: "Loop starten",
    valid: "Gültig",
    custom: "Benutzerdefinierten Block erstellen",
    failure: "Fehlerbehandlung",
    ai: "AI-Assistent",
  },
  {
    locale: "es",
    title: "Área de trabajo de Loop",
    blocks: "Bloques",
    code: "Código",
    run: "Ejecutar loop",
    valid: "Válido",
    custom: "Crear bloque personalizado",
    failure: "Gestión de fallos",
    ai: "Asistente AI",
  },
  {
    locale: "fr",
    title: "Atelier Loop",
    blocks: "Blocs",
    code: "Code",
    run: "Lancer le loop",
    valid: "Valide",
    custom: "Créer un bloc personnalisé",
    failure: "Gestion des échecs",
    ai: "Assistant IA",
  },
  {
    locale: "ja",
    title: "Loopワークベンチ",
    blocks: "ブロック",
    code: "コード",
    run: "Loopを実行",
    valid: "有効",
    custom: "カスタムブロックを作成",
    failure: "失敗処理",
    ai: "AIアシスタント",
  },
  {
    locale: "ko",
    title: "Loop 워크벤치",
    blocks: "블록",
    code: "코드",
    run: "Loop 실행",
    valid: "유효",
    custom: "사용자 블록 만들기",
    failure: "실패 처리",
    ai: "AI 어시스턴트",
  },
  {
    locale: "pt",
    title: "Bancada de Loop",
    blocks: "Blocos",
    code: "Código",
    run: "Executar loop",
    valid: "Válido",
    custom: "Criar bloco personalizado",
    failure: "Tratamento de falhas",
    ai: "Assistente de IA",
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

  test("creates custom blocks without putting Worker into the Loop AST", async ({
    page,
  }) => {
    await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);
    await expect(
      page.getByRole("heading", { name: "循环工作台" }),
    ).toBeVisible();
    await resetLoopSource(page, {
      blocks: "积木",
      code: "代码",
      run: "运行循环",
      valid: "有效",
    });

    await doubleClickBlocklyBackground(page);
    await page.getByRole("button", { name: "创建自定义积木" }).click();
    const dialog = page.getByRole("dialog", { name: "创建自定义积木" });
    await dialog.getByLabel("积木名称").fill("专业 PPT");
    await dialog.getByLabel("积木标识").fill("ppt-step");
    await dialog.getByLabel("任务模板").fill("制作 {{topic}} 的专业 PPT");
    await dialog.getByLabel("验证命令模板").fill("test -f {{file}}");
    await dialog.getByLabel("验收说明模板").fill("{{file}} 存在且可打开");
    await dialog.getByRole("button", { name: "创建积木" }).click();

    await expect(page.getByText("有效").first()).toBeVisible();
    await doubleClickBlocklyBackground(page);
    await expect(page.getByRole("button", { name: "专业 PPT" })).toBeVisible();
    await page.getByRole("button", { name: "专业 PPT" }).click();

    const codeEditor = page.locator(".cm-content");
    await expect(page.getByText("专业 PPT", { exact: true })).toBeVisible();
    await expect(page.getByText("topic", { exact: true })).toBeVisible();
    await expect(page.getByText("file", { exact: true })).toBeVisible();

    await expect(codeEditor).toContainText("agent ppt-step-task");
    const source = await codeEditor.innerText();
    expect(source.toLowerCase()).not.toContain("worker");
    expect(source).not.toContain("invalid-block-structure");
    expect(source).toContain("agent ppt-step-task");
    expect(source).toContain("verify ppt-step-check");
  });

  test("keeps Loop projection equivalent across eight localized workbenches", async ({
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

  test("keeps eight localized Loop workbenches usable on mobile", async ({
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

async function openLocalizedLoopWorkbench(
  page: Page,
  context: BrowserContext,
  localeCase: (typeof LOCALE_CASES)[number],
) {
  await context.addCookies([
    {
      name: LOCALE_COOKIE,
      value: localeCase.locale,
      url: getWebBaseUrl(),
    },
  ]);
  await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);
}

function collectFailedRequests(page: Page) {
  const failed: string[] = [];
  page.on("requestfailed", (request) => {
    if (isExpectedNavigationAbort(request)) return;
    failed.push(`${request.method()} ${request.url()} ${request.failure()?.errorText}`);
  });
  return () => failed;
}

function isExpectedNavigationAbort(request: Request) {
  const error = request.failure()?.errorText;
  if (error !== "net::ERR_ABORTED") return false;
  return /EventsService\/Subscribe|GoalLoopService\/CompileLoopProgram/.test(
    request.url(),
  );
}

async function resetLoopSource(
  page: Page,
  labels: { blocks: string; code: string; run: string; valid: string },
) {
  await page.getByRole("tab", { name: labels.code }).click();
  const codeEditor = page.locator(".cm-content");
  await expect(codeEditor).toBeEditable();
  await codeEditor.fill(DEFAULT_LOOP_SOURCE);
  await expect(page.getByRole("button", { name: labels.run })).toBeEnabled();
  await expect(page.getByText(labels.valid).first()).toBeVisible();
  await page.getByRole("tab", { name: labels.blocks }).click();
  await expect(page.locator(".blocklyMainBackground")).toBeVisible();
}

async function doubleClickBlocklyBackground(page: Page) {
  await page.locator(".blocklyMainBackground").first().waitFor();
  const point = await page.evaluate(() => {
    const background = document.querySelector(".blocklyMainBackground");
    if (!background) return undefined;
    const rect = background.getBoundingClientRect();
    const xSteps = [0.82, 0.72, 0.62, 0.52, 0.42, 0.32];
    const ySteps = [0.22, 0.34, 0.46, 0.58, 0.7, 0.82];
    for (const yStep of ySteps) {
      for (const xStep of xSteps) {
        const x = rect.left + rect.width * xStep;
        const y = rect.top + rect.height * yStep;
        if (document.elementFromPoint(x, y) === background) return { x, y };
      }
    }
    return undefined;
  });
  if (!point) throw new Error("No empty Blockly background point is available");
  await page.mouse.dblclick(point.x, point.y);
}

async function overflowingText(
  page: Page,
  labels: string[],
) {
  return page.evaluate((expectedLabels) => {
    return Array.from(
      document.querySelectorAll("button,[role='tab'],h1,h2,h3"),
    )
      .filter((element) => {
        const text = element.textContent?.trim() ?? "";
        return expectedLabels.some((label) => text.includes(label));
      })
      .map((element) => ({
        text: element.textContent?.trim() ?? "",
        overflows:
          element.scrollWidth > Math.ceil(element.getBoundingClientRect().width),
      }))
      .filter(({ overflows }) => overflows);
  }, labels);
}
