import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Create Worker UI", () => {
  test("shows quick creation and validates a missing task", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);

    await expect(page.getByRole("heading", {
      name: /What should this Worker do|这个 Worker 要做什么/i,
    })).toBeVisible();

    const createButton = page.getByRole("button", {
      name: /^(Create Worker|创建 Worker)$/i,
    });
    await expect(createButton).toBeEnabled();
    await createButton.click();
    await expect(page.getByText(
      /Describe the task before creating a Worker|请先描述任务，再创建 Worker/i,
    )).toBeVisible();
  });
});
