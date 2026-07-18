import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  ensureResourceWorkflowFixture,
  resetResourceWorkflowFixture,
} from "../../helpers/resource-workflow-fixture";

test.describe("Workflow Operations", () => {
  test.beforeEach(async ({ db }) => {
    clearAuthRateLimit();
    ensureResourceWorkflowFixture(db);
    resetResourceWorkflowFixture(db);
  });

  test("workflows: opens a new immutable revision editor", async ({
    page,
    db,
  }) => {
    const fixture = ensureResourceWorkflowFixture(db);
    await page.goto(`/${TEST_ORG_SLUG}/workflows/${fixture.slug}`);
    await page.getByRole("button", {
      name: /New revision|新建修订/i,
    }).click();

    await expect(page.getByTestId("resource-editor")).toBeVisible();
    await expect(page.getByLabel(/Resource name|资源名称/i))
      .toHaveValue(fixture.slug);
  });

  test("workflows: sidebar opens a resource-managed detail", async ({
    page,
    db,
  }) => {
    const fixture = ensureResourceWorkflowFixture(db);
    await page.goto(`/${TEST_ORG_SLUG}/workflows`);

    await page.locator(
      `[data-testid="workflow-row"][data-workflow-slug="${fixture.slug}"]`,
    ).click();

    await expect(page).toHaveURL(
      new RegExp(`/${TEST_ORG_SLUG}/workflows/${fixture.slug}$`),
    );
    await expect(page.getByRole("heading", {
      level: 1,
      name: fixture.name,
    })).toBeVisible();
  });
});
