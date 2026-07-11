import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

/**
 * Workflow ↔ EnvBundle binding end-to-end.
 *
 * Covers the I4 contract: creating a Workflow with `usedEnvBundles = ["<name>", ...]`
 * persists the ordered list, GetWorkflow round-trips it, UpdateWorkflow clears it via
 * empty list, and unknown-bundle names still create the Workflow (eval is warn-only
 * at run-time).
 *
 * Both Workflow CRUD and EnvBundle CRUD live on Connect-RPC (R6 completion).
 *
 * Pod-level KV injection is left to higher-tier integration tests since it
 * requires a Pod to actually launch and read its env.
 */
test.describe("Workflow ↔ EnvBundle binding", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  test("Workflow persists usedEnvBundles (multi) and round-trips on GetWorkflow", async ({ api }) => {
    const cc = await api.connect();
    const ts = Date.now();
    const bundleAName = `e2e-workflow-A-${ts}`;
    const bundleBName = `e2e-workflow-B-${ts}`;

    const createBundle = async (name: string) =>
      cc.envBundle.createEnvBundle({
        agentSlug: "claude-code",
        name,
        kind: "credential",
        data: { ANTHROPIC_API_KEY: "sk-test-e2e" },
      }) as Promise<{ id: bigint }>;

    const bundleA = await createBundle(bundleAName);
    const bundleB = await createBundle(bundleBName);
    expect(bundleA.id).toBeTruthy();
    expect(bundleB.id).toBeTruthy();

    let slug: string | undefined;
    try {
      const created = await cc.workflow.createWorkflow({
        orgSlug: TEST_ORG_SLUG,
        name: `E2E Workflow Bundle ${ts}`,
        agentSlug: "claude-code",
        promptTemplate: "echo bound",
        usedEnvBundles: [bundleAName, bundleBName],
      }) as { slug: string; usedEnvBundles: string[] };
      slug = created.slug;
      expect(slug).toBeTruthy();
      // Order preserved exactly.
      expect(created.usedEnvBundles).toEqual([bundleAName, bundleBName]);

      const fetched = await cc.workflow.getWorkflow({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: slug,
      }) as { usedEnvBundles: string[] };
      expect(fetched.usedEnvBundles).toEqual([bundleAName, bundleBName]);
    } finally {
      if (slug) {
        await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: slug }).catch(() => null);
      }
      await cc.envBundle.deleteEnvBundle({ id: bundleA.id }).catch(() => null);
      await cc.envBundle.deleteEnvBundle({ id: bundleB.id }).catch(() => null);
    }
  });

  test("UpdateWorkflow with usedEnvBundles={names:[]} clears the binding", async ({ api }) => {
    const cc = await api.connect();
    const ts = Date.now();
    const bundleName = `e2e-clear-${ts}`;

    const bundle = await cc.envBundle.createEnvBundle({
      agentSlug: "claude-code",
      name: bundleName,
      kind: "credential",
      data: { ANTHROPIC_API_KEY: "sk-test-e2e-clear" },
    }) as { id: bigint };
    expect(bundle.id).toBeTruthy();

    let slug: string | undefined;
    try {
      const created = await cc.workflow.createWorkflow({
        orgSlug: TEST_ORG_SLUG,
        name: `E2E Workflow Clear ${ts}`,
        agentSlug: "claude-code",
        promptTemplate: "echo bound",
        usedEnvBundles: [bundleName],
      }) as { slug: string };
      slug = created.slug;
      expect(slug).toBeTruthy();

      // Wrapper present with empty `names` explicitly clears the binding.
      await cc.workflow.updateWorkflow({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: slug,
        usedEnvBundles: { names: [] },
      });

      const after = await cc.workflow.getWorkflow({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: slug,
      }) as { usedEnvBundles: string[] };
      // Backend returns [] (not null) for an empty array column.
      expect(after.usedEnvBundles).toEqual([]);
    } finally {
      if (slug) {
        await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: slug }).catch(() => null);
      }
      await cc.envBundle.deleteEnvBundle({ id: bundle.id }).catch(() => null);
    }
  });

  test("Workflow with unknown bundle name is still creatable (warn-only at run-time)", async ({ api }) => {
    const cc = await api.connect();
    const ts = Date.now();
    // Use a name we know does NOT exist; the AgentFile eval contract is
    // tolerant of dangling references (USE_ENV_BUNDLE skips silently when
    // the name isn't in ctx.EnvBundles).
    const created = await cc.workflow.createWorkflow({
      orgSlug: TEST_ORG_SLUG,
      name: `E2E Workflow Dangling ${ts}`,
      agentSlug: "claude-code",
      promptTemplate: "echo dangling",
      usedEnvBundles: [`nonexistent-bundle-${ts}`],
    }) as { slug: string };
    const slug = created.slug;
    expect(slug).toBeTruthy();
    await cc.workflow.deleteWorkflow({ orgSlug: TEST_ORG_SLUG, workflowSlug: slug }).catch(() => null);
  });
});
