// Multi-tab UI propagation for workflow_run:started.
//
// Both tabs open the same workflow detail page; tab A triggers a run via
// Connect-RPC and tab B's run-history list grows by 1 without reload.
//
// Wire-level coverage in tests/realtime/workflow-events-wire.spec.ts; this
// spec exercises handler → fetchRuns → React render chain.
import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { TEST_ORG_SLUG } from "../../helpers/env";
import {
  ensureResourceWorkflowFixture,
  resetResourceWorkflowFixture,
} from "../../helpers/resource-workflow-fixture";

test.describe("Workflow run · multi-tab UI propagation", () => {
  test.beforeEach(async ({ db }) => {
    clearAuthRateLimit();
    await terminateAllPods();
    ensureResourceWorkflowFixture(db);
    resetResourceWorkflowFixture(db);
  });
  test.afterEach(async ({ db }) => {
    await terminateAllPods();
    resetResourceWorkflowFixture(db);
  });

  test("tab A trigger run → tab B run-history list adds card", async ({
    context,
    api,
    db,
  }) => {
    const cc = await api.connect();
    const workflow = ensureResourceWorkflowFixture(db);

    const tabA = await context.newPage();
    const tabB = await context.newPage();
    await Promise.all([
      tabA.goto(`/${TEST_ORG_SLUG}/workflows/${workflow.slug}`),
      tabB.goto(`/${TEST_ORG_SLUG}/workflows/${workflow.slug}`),
    ]);

    // Wait until the WorkflowHeader's h1 renders the workflow name in BOTH tabs —
    // that means fetchWorkflow resolved, currentWorkflow is set in WASM, and the
    // realtime handler (which reads currentWorkflow synchronously) will route
    // the upcoming workflow_run:started event correctly.
    await Promise.all([
      expect(tabA.getByRole("heading", {
        level: 1,
        name: workflow.name,
      })).toBeVisible({ timeout: 30_000 }),
      expect(tabB.getByRole("heading", {
        level: 1,
        name: workflow.name,
      })).toBeVisible({ timeout: 30_000 }),
    ]);

    // EventSubscriptionManager bootstrap window so both tabs are subscribed
    // before publish. Events published before subscribeAll registers have
    // no replay buffer.
    await tabA.waitForTimeout(2000);

    const runCard = `[data-testid="workflow-run-card"]`;
    // Workflow just created: no runs yet — assert no cards mounted.
    await Promise.all([
      expect(tabA.locator(runCard)).toHaveCount(0),
      expect(tabB.locator(runCard)).toHaveCount(0),
    ]);

    await cc.workflow.triggerWorkflow({
      orgSlug: TEST_ORG_SLUG, workflowSlug: workflow.slug,
    } as never);

    // workflow_run:started → debounced refetch (500ms) → fetchRuns. Both tabs
    // should observe at least 1 run-card within the window.
    await Promise.all([
      expect(tabA.locator(runCard)).toHaveCount(1, { timeout: 15_000 }),
      expect(tabB.locator(runCard)).toHaveCount(1, { timeout: 15_000 }),
    ]);

    await tabA.close();
    await tabB.close();
  });
});
