import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  ensureResourceWorkflowFixture,
} from "../../helpers/resource-workflow-fixture";
import { cleanupResourceWorkflowFixture } from "../../helpers/resource-workflow-run-cleanup";

test.describe("Workflow Operations", () => {
  test.beforeEach(async ({ db, api }) => {
    clearAuthRateLimit();
    await cleanupResourceWorkflowFixture(db, api);
    await ensureResourceWorkflowFixture(db, api);
  });
  test.afterEach(async ({ db, api }) => {
    await cleanupResourceWorkflowFixture(db, api);
  });

  test("workflows: opens a new immutable revision editor", async ({
    page,
    db,
    api,
  }) => {
    const fixture = await ensureResourceWorkflowFixture(db, api);
    await page.goto(`/${TEST_ORG_SLUG}/workflows/${fixture.slug}`);
    await page.getByRole("button", {
      name: /^New revision$|^新建修订$/,
    }).click();

    await expect(page.getByTestId("resource-editor")).toBeVisible();
    await expect(page.getByLabel(/Resource name|资源名称/i))
      .toHaveValue(fixture.slug);
  });

  test("workflows: sidebar opens a resource-managed detail", async ({
    page,
    db,
    api,
  }) => {
    const fixture = await ensureResourceWorkflowFixture(db, api);
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
