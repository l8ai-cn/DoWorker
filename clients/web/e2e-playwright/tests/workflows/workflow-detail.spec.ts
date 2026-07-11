// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
test.describe("Workflow Detail Page", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  let createdSlug: string | null = null;

  test.afterEach(async ({ api }) => {
    if (createdSlug) {
      const cc = await api.connect();
      await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: createdSlug }).catch(() => null);
      createdSlug = null;
    }
  });

  test("API: get workflow detail returns entity", async ({ api }) => {
    const cc = await api.connect();
    const created = await cc.workflow.createWorkflow({
      orgSlug: TEST_ORG_SLUG,
      name: `E2E Workflow Detail ${Date.now()}`,
      slug: `e2e-workflow-detail-${Date.now()}`,
      agentSlug: "claude-code",
      cronExpression: "0 * * * *",
      promptTemplate: "echo test",
    }) as { slug: string };
    createdSlug = created.slug;

    const workflow = await cc.workflow.getWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: createdSlug }) as { slug: string };
    expect(workflow.slug).toBe(createdSlug);
  });

  test("UI: workflow detail page renders without errors", async ({ page, api }) => {
    const cc = await api.connect();
    const created = await cc.workflow.createWorkflow({
      orgSlug: TEST_ORG_SLUG,
      name: `E2E Workflow UI ${Date.now()}`,
      slug: `e2e-workflow-ui-${Date.now()}`,
      agentSlug: "claude-code",
      cronExpression: "0 * * * *",
      promptTemplate: "echo test",
    }) as { slug: string };
    createdSlug = created.slug;
    await page.goto(`/${TEST_ORG_SLUG}/workflows/${createdSlug}`);
    await page.waitForLoadState("load");
  });
});
