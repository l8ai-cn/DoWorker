import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  ensureResourceWorkflowFixture,
  resetResourceWorkflowFixture,
} from "../../helpers/resource-workflow-fixture";

test.describe("Workflow Detail Page", () => {
  test.beforeEach(async ({ db }) => {
    clearAuthRateLimit();
    ensureResourceWorkflowFixture(db);
    resetResourceWorkflowFixture(db);
  });

  test("API: get workflow detail returns entity", async ({ api, db }) => {
    const cc = await api.connect();
    const fixture = ensureResourceWorkflowFixture(db);
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
  }) => {
    const fixture = ensureResourceWorkflowFixture(db);
    await page.goto(`/${TEST_ORG_SLUG}/workflows/${fixture.slug}`);

    await expect(page.getByRole("heading", {
      level: 1,
      name: fixture.name,
    })).toBeVisible();
    await expect(page.getByText("resource-native")).toBeVisible();
  });
});
