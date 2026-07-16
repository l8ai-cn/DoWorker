import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

test.describe("Create Worker UI", () => {
  test("shows the resource editor and prevents incomplete creation", async ({ page }) => {
    await page.goto(`/${TEST_ORG_SLUG}/workers/new`);
    const editor = page.getByTestId("resource-editor");

    await expect(editor).toBeVisible();
    await expect(editor.getByRole("button", {
      name: /Create Worker|创建 Worker/i,
    })).toBeDisabled();
  });
});
