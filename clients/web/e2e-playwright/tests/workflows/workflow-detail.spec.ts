import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  ensureResourceWorkflowFixture,
} from "../../helpers/resource-workflow-fixture";
import { cleanupResourceWorkflowFixture } from "../../helpers/resource-workflow-run-cleanup";

test.describe("Workflow Detail Page", () => {
  test.beforeEach(async ({ db, api }) => {
    clearAuthRateLimit();
    await cleanupResourceWorkflowFixture(db, api);
    await ensureResourceWorkflowFixture(db, api);
  });
  test.afterEach(async ({ db, api }) => {
    await cleanupResourceWorkflowFixture(db, api);
  });

  test("API: get workflow detail returns entity", async ({ api, db }) => {
    const cc = await api.connect();
    const fixture = await ensureResourceWorkflowFixture(db, api);
    const workflow = await cc.workflow.getWorkflow({
      orgSlug: TEST_ORG_SLUG,
      workflowSlug: fixture.slug,
    }) as { agentSlug: string; slug: string };

    expect(workflow.slug).toBe(fixture.slug);
    expect(workflow.agentSlug).toBe("resource-native");
  });

  test("UI: workflow detail page renders resource-managed entity", async ({
    page,
    db,
    api,
  }) => {
    const fixture = await ensureResourceWorkflowFixture(db, api);
    await page.goto(`/${TEST_ORG_SLUG}/workflows/${fixture.slug}`);

    await expect(page.getByRole("heading", {
      level: 1,
      name: fixture.name,
    })).toBeVisible();
    await expect(page.getByText("resource-native")).toBeVisible();
  });
});
