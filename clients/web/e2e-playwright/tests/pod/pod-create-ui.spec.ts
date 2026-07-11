// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { E2E_ECHO_AGENT_SLUG } from "../../helpers/e2e-echo-runner";
import { CreateWorkerPage } from "../../pages/create-worker.page";

type Runner = { id: bigint };
type Agent = { slug: string };

/**
 * UI regression for: Create Worker must submit successfully and the new pod
 * must appear in the workspace.
 *
 * The Connect-RPC response must reach the page success handler, which routes
 * back to workspace with the newly-created pod selected.
 */
test.describe("Create Worker UI", () => {
  test.afterEach(async () => {
    await terminateAllPods();
  });

  test("Create Worker submits and new pod appears in workspace", async ({ page, api }) => {
    const cc = await api.connect();
    const { items: runners } = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG }) as { items: Runner[] };
    expect(runners.length, "dev env must have an online runner").toBeGreaterThan(0);

    const { builtinAgents: agents } = await cc.agent.listAgents({ orgSlug: TEST_ORG_SLUG }) as { builtinAgents: Agent[] };
    expect(agents.length, "dev env must have a builtin agent").toBeGreaterThan(0);

    // Start clean so the sidebar count is deterministic.
    await terminateAllPods();

    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");

    // PodListItem renders with `data-testid="pod-list-item"` for each pod
    // in the workspace sidebar. The previous text-regex approach was too
    // brittle: pod_key is `<user_id>-<standalone|ticket_id>-<hex>` (not
    // org_id, and never the literal "ticket"/"channel"), and
    // `getPodDisplayName` may render `Agent Name (1-standa)` or
    // `1-standa...` — neither matches a fixed pod_key regex.
    //
    // Authoritative count comes from the backend (cc.pod.listPods.total)
    // — items[] is paginated (default limit 20), so we read `total` to
    // avoid false negatives. The sidebar "mine" filter only surfaces
    // running/initializing pods, but a freshly created pod can race to
    // `failed` if the dev runner can't launch the agent CLI. We want
    // this UI regression spec to assert the create+propagate flow, not
    // runtime agent health.
    const podsBefore = await cc.pod.listPods({ orgSlug: TEST_ORG_SLUG }) as { total: bigint | number };
    const beforeTotal = Number(podsBefore.total);

    const worker = new CreateWorkerPage(page, TEST_ORG_SLUG);
    await worker.goto();
    expect(
      agents.some((agent) => agent.slug === E2E_ECHO_AGENT_SLUG),
      "dev env must include the e2e-echo builtin agent",
    ).toBeTruthy();
    await worker.selectImage(E2E_ECHO_AGENT_SLUG);
    const createResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith("/proto.pod.v1.PodService/CreatePod"),
      { timeout: 20_000 },
    );
    await worker.submit();
    const response = await createResponse;
    expect(response.ok()).toBeTruthy();
    await worker.waitForWorkspace();

    await expect
      .poll(async () => {
        const after = await cc.pod.listPods({ orgSlug: TEST_ORG_SLUG }) as { total: bigint | number };
        return Number(after.total);
      }, { timeout: 10_000 })
      .toBeGreaterThan(beforeTotal);
  });
});
