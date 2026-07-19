import type { BrowserContext, Page, Request } from "@playwright/test";
import { expect } from "../../fixtures/index";
import { getWebBaseUrl, TEST_ORG_SLUG } from "../../helpers/env";
import { LOCALE_COOKIE } from "../../../src/lib/i18n/config";

export interface LocalizedLoopWorkbench {
  locale: string;
  title: string;
  blocks: string;
  code: string;
  run: string;
  valid: string;
}

export function collectFailedRequests(page: Page) {
  const failed: string[] = [];
  page.on("requestfailed", (request) => {
    if (isExpectedNavigationAbort(request)) return;
    failed.push(`${request.method()} ${request.url()} ${request.failure()?.errorText}`);
  });
  return () => failed;
}

export async function doubleClickBlocklyBackground(page: Page) {
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

export async function openLocalizedLoopWorkbench(
  page: Page,
  context: BrowserContext,
  localeCase: LocalizedLoopWorkbench,
) {
  await context.addCookies([{
    name: LOCALE_COOKIE,
    value: localeCase.locale,
    url: getWebBaseUrl(),
  }]);
  await page.goto(`/${TEST_ORG_SLUG}/loops/workbench`);
}

export async function overflowingText(page: Page, labels: string[]) {
  return page.evaluate((expectedLabels) => {
    return Array.from(document.querySelectorAll("button,[role='tab'],h1,h2,h3"))
      .filter((element) => {
        const text = element.textContent?.trim() ?? "";
        return expectedLabels.some((label) => text.includes(label));
      })
      .map((element) => ({
        text: element.textContent?.trim() ?? "",
        overflows: element.scrollWidth > Math.ceil(element.getBoundingClientRect().width),
      }))
      .filter(({ overflows }) => overflows);
  }, labels);
}

export async function resetLoopSource(
  page: Page,
  labels: { blocks: string; code: string; run: string; valid: string },
) {
  await page.getByRole("tab", { name: labels.code }).click();
  const codeEditor = page.locator(".cm-content");
  await expect(codeEditor).toBeEditable();
  await codeEditor.fill(`@id(n-checkout-fix)
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
}`);
  await expect(page.getByRole("button", { name: labels.run })).toBeEnabled();
  await expect(page.getByText(labels.valid).first()).toBeVisible();
  await page.getByRole("tab", { name: labels.blocks }).click();
  await expect(page.locator(".blocklyMainBackground")).toBeVisible();
}

function isExpectedNavigationAbort(request: Request) {
  const error = request.failure()?.errorText;
  return error === "net::ERR_ABORTED" &&
    /EventsService\/Subscribe|GoalLoopService\/CompileLoopProgram/.test(request.url());
}
