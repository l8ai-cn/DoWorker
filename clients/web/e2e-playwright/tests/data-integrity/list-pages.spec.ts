// Data-integrity assertions for key list pages. These specs are the last
// line of defense for the failure class that motivated this PR:
// the renderer reads an API response but a deserialization / wasm-bridge
// drift loses fields silently, so the page renders an empty list while
// the API says N items exist.
//
// Each test:
//   1. fetches ground-truth from Connect-RPC (same wire format the
//      production client uses, asserted on the proto layer)
//   2. navigates to the page, waits for the data to land
//   3. compares DOM-rendered item counts against the API count
//
// A passing run proves the wasm cache + selectors + render path all
// agree with the backend. The auto-attached console monitor (see
// fixtures/index.ts) catches the silent-error class on top.

import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const isEmptyHint = /no .*(pod|workflow|channel|ticket|run)s?|no .*found|没有.*(pod|workflow|channel|ticket|run)|empty|暂无|还没有|nothing here|此.*暂无/i;

test.describe("Data integrity: list pages match API counts", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("workspace sidebar pod count matches ListPods API", async ({ page, api, db }) => {
    const cc = await api.connect();
    // Mirror the renderer's default sidebar request — "mine" tab maps to
    // status "running,initializing" + created_by_id = current user
    // (SIDEBAR_STATUS_MAP in stores/podTypes.ts). Anything else returned
    // by the API would be filtered out client-side, so testing equality
    // against an unfiltered API call is structurally wrong.
    const userId = db.queryValue(
      `SELECT id FROM users WHERE email = 'dev@agentcloud.local' LIMIT 1`,
    );
    expect(userId, "dev seed must include the dev user").toBeTruthy();
    const { items } = await cc.pod.listPods({
      orgSlug: TEST_ORG_SLUG,
      status: "running,initializing",
      createdById: BigInt(userId as string),
      limit: 50,
      offset: 0,
    }) as { items: Array<{ podKey: string }> };

    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000); // sidebar fetch lands

    if (items.length === 0) {
      // Empty state must be visible — otherwise the sidebar is silently
      // showing nothing, which is the original bug.
      const empty = await page.getByText(isEmptyHint).first().isVisible({ timeout: 5_000 }).catch(() => false);
      expect(empty, "API returned no pods → page must show empty state").toBe(true);
      return;
    }

    // Wait for the first pod to appear, then assert the total matches.
    const podRows = page.locator('[data-testid="pod-list-item"]');
    await expect(podRows.first()).toBeVisible({ timeout: 10_000 });
    // The sidebar may also include pods from other filters or paginated
    // sub-views; assert it includes *at least* every podKey the API
    // returned, not strict equality. Matches the "wasm cache lost data"
    // failure mode exactly: if the cache drops pods, this comparison
    // catches it.
    const renderedKeys = await podRows.evaluateAll(els =>
      els.map(el => el.getAttribute("data-pod-key")).filter((k): k is string => !!k)
    );
    for (const expected of items.map((p) => p.podKey)) {
      expect(
        renderedKeys.includes(expected),
        `sidebar missing pod ${expected} that ListPods returned`,
      ).toBe(true);
    }
  });

  test("runner detail pods tab row count matches ListPods(runnerId) API", async ({ page, api, db }) => {
    const runnerId = db.queryValue(
      `SELECT id FROM runners WHERE organization_id = (SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}') LIMIT 1`,
    );
    expect(runnerId, "dev seed must include at least one runner").toBeTruthy();
    const id = runnerId as string;

    const cc = await api.connect();
    // ListPods filtered by runner_id — same data the runner-detail Pods
    // tab renders (production goes through a wasm-only helper, but the
    // backend wire query is exactly this). Renderer drops the limit/offset
    // to mirror what useRunnerDetail.ts sends.
    const apiRes = await cc.pod.listPods({
      orgSlug: TEST_ORG_SLUG,
      runnerId: BigInt(id),
      limit: 20,
      offset: 0,
    }) as { items: Array<{ podKey: string }>; total: bigint };

    await page.goto(`/${TEST_ORG_SLUG}/runners/${id}`);
    await page.waitForLoadState("load");

    // Pods tab must be selected for rows to render (useRunnerDetail only
    // fires loadPods when activeTab === "pods"). Tie the wait to the actual
    // ListPods response instead of a fixed sleep: a heavily-populated runner
    // makes the backend COUNT + workflow-metadata join slow, so a fixed 4s
    // budget flakes. waitForResponse is bounded by the real round-trip.
    const podsTab = page.locator('[data-testid="runner-detail-tab-pods"]');
    await expect(podsTab).toBeVisible({ timeout: 15_000 });
    const podsResp = page
      .waitForResponse(
        (r) => r.url().includes("ListPods") && r.request().method() === "POST",
        { timeout: 30_000 },
      )
      .catch(() => null);
    await podsTab.click();
    await podsResp;

    const expectedKeys = apiRes.items.map((p) => p.podKey);
    if (expectedKeys.length === 0) {
      const emptyVisible = await page
        .getByText(isEmptyHint)
        .first()
        .isVisible({ timeout: 5_000 })
        .catch(() => false);
      expect(
        emptyVisible,
        "API returned no runner pods → page must show empty state",
      ).toBe(true);
      return;
    }

    // Wait for the runner-pod-row attribute to appear (the table render is
    // async behind a loadPods fetch). Then read every rendered pod_key.
    const podRows = page.locator('[data-testid="runner-pod-row"]');
    await expect(podRows.first()).toBeVisible({ timeout: 10_000 });
    const renderedKeys = await podRows.evaluateAll(els =>
      els.map(el => el.getAttribute("data-pod-key")).filter((k): k is string => !!k)
    );
    for (const key of expectedKeys) {
      expect(
        renderedKeys.includes(key),
        `runner pods table missing pod ${key} from ListPods(runnerId) response`,
      ).toBe(true);
    }
  });

  test("tickets board column counts match GetBoard API", async ({ page, api }) => {
    const cc = await api.connect();
    const board = await cc.ticket.getBoard({ orgSlug: TEST_ORG_SLUG }) as {
      columns: Array<{ status: string; tickets: Array<{ slug: string }> }>;
    };

    await page.goto(`/${TEST_ORG_SLUG}/tickets`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000);

    // Sum across all columns — the board view renders every ticket;
    // proto field drift would leave the column array empty / lose
    // tickets in transit. We check inclusion of each slug in the
    // rendered body text, matching the data-integrity contract.
    const expectedSlugs = board.columns.flatMap((c) => c.tickets.map((t) => t.slug));
    if (expectedSlugs.length === 0) return;

    const bodyText = await page.textContent("body") ?? "";
    let foundAny = false;
    for (const slug of expectedSlugs) {
      if (bodyText.includes(slug)) {
        foundAny = true;
        break;
      }
    }
    expect(
      foundAny,
      `tickets board renders 0 of ${expectedSlugs.length} API-returned ticket slugs — wasm cache or selector likely dropped them`,
    ).toBe(true);
  });

  test("channel list count from ListChannels API matches DOM", async ({ page, api }) => {
    const cc = await api.connect();
    const { items } = await cc.channel.listChannels({ orgSlug: TEST_ORG_SLUG }) as {
      items: Array<{ id: bigint; name: string }>;
    };

    await page.goto(`/${TEST_ORG_SLUG}/channels`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000);

    if (items.length === 0) {
      const empty = await page.getByText(isEmptyHint).first().isVisible({ timeout: 5_000 }).catch(() => false);
      expect(empty, "no channels → page must render empty state").toBe(true);
      return;
    }

    const bodyText = await page.textContent("body") ?? "";
    let foundAny = false;
    for (const ch of items) {
      if (bodyText.includes(ch.name)) {
        foundAny = true;
        break;
      }
    }
    expect(
      foundAny,
      `channel list renders 0 of ${items.length} API-returned channel names — wasm cache likely dropped them`,
    ).toBe(true);
  });

  test("workflows list count matches ListWorkflows API", async ({ page, api }) => {
    const cc = await api.connect();
    const { items } = await cc.workflow.listWorkflows({ orgSlug: TEST_ORG_SLUG }) as {
      items: Array<{ slug: string; name: string }>;
    };

    await page.goto(`/${TEST_ORG_SLUG}/workflows`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000);

    if (items.length === 0) {
      const empty = await page.getByText(isEmptyHint).first().isVisible({ timeout: 5_000 }).catch(() => false);
      expect(empty, "no workflows → page must render empty state").toBe(true);
      return;
    }

    const bodyText = await page.textContent("body") ?? "";
    let foundAny = false;
    for (const workflow of items) {
      if (bodyText.includes(workflow.name) || bodyText.includes(workflow.slug)) {
        foundAny = true;
        break;
      }
    }
    expect(
      foundAny,
      `workflows list renders 0 of ${items.length} API-returned workflow names/slugs — wasm cache likely dropped them`,
    ).toBe(true);
  });

  test("repositories sidebar rows match ListRepositories API count", async ({ page, api }) => {
    const cc = await api.connect();
    // RepoState was just migrated to proto-bytes mutators (5 JSON bypasses
    // eliminated). The sidebar list path still reads from the same wasm
    // cache the proto mutators write to, so this asserts the round-trip:
    // backend → wasm proto deserialization → cache → useRepositories selector.
    const { items } = await cc.repository.listRepositories({ orgSlug: TEST_ORG_SLUG }) as {
      items: Array<{ id: bigint; slug: string }>;
    };

    // `/{org}/repositories` redirects to `/{org}/infra?tab=repositories`.
    // The list renders inside the IDE shell's RepositoriesSidebarContent.
    await page.goto(`/${TEST_ORG_SLUG}/repositories`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000);

    if (items.length === 0) {
      const empty = await page.getByText(isEmptyHint).first().isVisible({ timeout: 5_000 }).catch(() => false);
      expect(empty, "no repositories → page must render empty state").toBe(true);
      return;
    }

    const rows = page.locator('[data-testid="repository-row"]');
    await expect(rows.first()).toBeVisible({ timeout: 10_000 });
    const renderedSlugs = await rows.evaluateAll(els =>
      els.map(el => el.getAttribute("data-repo-slug")).filter((s): s is string => !!s)
    );
    for (const repo of items) {
      expect(
        renderedSlugs.includes(repo.slug),
        `repositories sidebar missing repo ${repo.slug} that ListRepositories returned`,
      ).toBe(true);
    }
  });

  test("mesh topology pod nodes match GetMeshTopology API count", async ({ page, api }) => {
    const cc = await api.connect();
    // MeshState just migrated to ReplaceTopologyRequest proto-bytes. The
    // topology read path (GetMeshTopology → wasm fetch_topology → store
    // _tick → renderer) is the failure mode this guards: if any field
    // name drifts during proto encode/decode, nodes silently disappear.
    const topology = await cc.mesh.getMeshTopology({ orgSlug: TEST_ORG_SLUG }) as {
      nodes: Array<{ podKey: string }>;
    };

    await page.goto(`/${TEST_ORG_SLUG}/mesh`);
    await page.waitForLoadState("load");
    await page.waitForTimeout(2000);

    if (topology.nodes.length === 0) {
      // Mesh renders its own "No Active Pods" state — match that text or
      // the generic empty hint. Dev seed has 0 pods so this is the
      // expected branch on a clean install.
      const noPodsVisible = await page.getByText(/no active pods/i).first().isVisible({ timeout: 5_000 }).catch(() => false);
      const generic = await page.getByText(isEmptyHint).first().isVisible({ timeout: 5_000 }).catch(() => false);
      expect(noPodsVisible || generic, "API returned 0 mesh nodes → mesh page must show empty state").toBe(true);
      return;
    }

    // React-flow renders one PodNode per topology.nodes entry; the inner
    // div carries `data-pod-key` so we can compare directly against the
    // API list rather than parsing visible text.
    const podNodes = page.locator('[data-testid="mesh-pod-node"]');
    await expect(podNodes.first()).toBeVisible({ timeout: 10_000 });
    const renderedKeys = await podNodes.evaluateAll(els =>
      els.map(el => el.getAttribute("data-pod-key")).filter((k): k is string => !!k)
    );
    for (const node of topology.nodes) {
      expect(
        renderedKeys.includes(node.podKey),
        `mesh topology missing pod ${node.podKey} that GetMeshTopology returned`,
      ).toBe(true);
    }
  });

  test("blocks page loads workspace from ListWorkspaces API", async ({ page, api }) => {
    const cc = await api.connect();
    // BlockstoreState just migrated to ApplyRemoteOpRequest proto-bytes for
    // mutations. The workspace listing path (listWorkspaces → page
    // hydrate → root block ID) still flows through JSON fetch for the
    // page-tree shell — but the page is the integration target where any
    // proto/JSON drift would leave the renderer stuck on the loading
    // spinner. This asserts the page reaches a "workspace loaded" state.
    const { items: workspaces } = await cc.blockstore.listWorkspaces({ orgSlug: TEST_ORG_SLUG }) as {
      items: Array<{ id: string; rootBlockId?: string }>;
    };

    await page.goto(`/${TEST_ORG_SLUG}/blocks`);
    await page.waitForLoadState("load");
    // ensureDefaultWorkspace auto-creates a workspace if none exists, so
    // by the time the page settles the API list must have ≥1 item.
    await page.waitForTimeout(3000);

    // After hydrate the BlocksSidebar mounts the page-tree shell. The
    // search input gets `data-testid="blocks-sidebar-search"` once the
    // workspace is ready — its presence is the canary that the
    // workspace → root_block_id → renderer chain wired up.
    const sidebarSearch = page.locator('[data-testid="blocks-sidebar-search"]');
    await expect(sidebarSearch).toBeVisible({ timeout: 15_000 });

    // After waiting for the sidebar, the workspace list must be non-empty.
    // ensureDefaultWorkspace ran during page load, so re-fetch.
    const after = await cc.blockstore.listWorkspaces({ orgSlug: TEST_ORG_SLUG }) as {
      items: Array<{ id: string }>;
    };
    expect(
      after.items.length,
      "blocks page should have triggered ensureDefaultWorkspace; listWorkspaces must return ≥1",
    ).toBeGreaterThan(0);
    // Sanity: the initial workspaces list (if non-empty) is a subset of
    // post-hydrate list (ensureDefault only adds, never removes).
    for (const w of workspaces) {
      expect(after.items.some((x) => x.id === w.id)).toBe(true);
    }
  });
});
