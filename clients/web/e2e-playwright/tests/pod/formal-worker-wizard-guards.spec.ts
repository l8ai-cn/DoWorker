import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

const imageBlockedWorkers = [
  "Aider",
  "Claude Code",
  "Cursor CLI",
  "DoAgent",
  "Grok Build",
  "Hermes",
  "Loopal",
  "MiniMax CLI",
  "OpenClaw",
  "OpenCode",
];

test.describe("Formal Worker wizard guards", () => {
  test("requires a Gemini model resource and exposes every missing-image block", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    const workerType = page.getByLabel(/^(Worker Type|Worker 类型)$/i);
    await workerType.click();
    for (const worker of imageBlockedWorkers) {
      await expect(page.getByRole("option", {
        name: new RegExp(`${worker}.*No runtime image is available`, "i"),
      })).toBeVisible();
    }

    await page.getByRole("option", { name: "Gemini CLI", exact: true }).click();
    const changeType = page.getByRole("button", {
      name: /^(Change type|Switch type|切换类型)$/i,
    });
    if (await changeType.isVisible()) await changeType.click();

    await expect(page.getByTestId("worker-runtime-field-model")).toContainText(
      /No compatible model resources|暂无兼容的模型资源/i,
    );
    await expect(page.getByRole("button", { name: /^(Next|下一步)$/i })).toBeDisabled();
  });
});
