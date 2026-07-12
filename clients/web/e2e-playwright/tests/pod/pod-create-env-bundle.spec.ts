import { test, expect } from "../../fixtures/index";
import { fromBinary } from "@bufbuild/protobuf";
import { CreatePodRequestSchema } from "../../../../../proto/gen/ts/pod/v1/pod_pb";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { clearAuthRateLimit } from "../../helpers/redis";
import { CreateWorkerPage } from "../../pages/create-worker.page";

/**
 * Pod creation × EnvBundle UI flow.
 *
 * The Create Pod modal renders runtime bundles as an ordered multi-select.
 * Model/API credentials are selected through AI Resources, not EnvBundles.
 *
 * We don't have a persisted `pods.agentfile_layer` column — the merged
 * layer is built per-request and shipped to Runner. So we verify the wire
 * contract via Playwright route interception: the Connect-RPC CreatePod
 * binary request carries the expected agentfile_layer with the expected
 * lines in the expected order.
 */
const CREATE_POD_RPC = "/proto.pod.v1.PodService/CreatePod";
const AGENT_SLUG = "e2e-echo";

function decodeCreatePodLayer(rawBody: Buffer | string | null): string | undefined {
  if (!rawBody) return undefined;
  const bytes =
    typeof rawBody === "string"
      ? new Uint8Array(Buffer.from(rawBody, "binary"))
      : new Uint8Array(rawBody);
  try {
    const msg = fromBinary(CreatePodRequestSchema, bytes);
    return msg.agentfileLayer;
  } catch {
    return undefined;
  }
}

test.describe("Pod create — EnvBundle binding UI", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  test.afterEach(async () => {
    await terminateAllPods();
  });

  test("Pod create dialog attaches selected runtime bundles in order", async ({
    page,
    api,
    db,
  }) => {
    const cc = await api.connect();
    const { items: runners } = await cc.runner.listAvailableRunners({
      orgSlug: TEST_ORG_SLUG,
    }) as { items: unknown[] };
    expect(runners.length, "dev env must have an online runner").toBeGreaterThan(0);
    const { builtinAgents } = await cc.agent.listAgents({
      orgSlug: TEST_ORG_SLUG,
    }) as { builtinAgents: { slug: string }[] };
    expect(
      builtinAgents?.some((a) => a.slug === AGENT_SLUG),
      "dev env must include the e2e-echo builtin agent",
    ).toBeTruthy();

    const stamp = Date.now();
    const runtimeName = `E2E PodUI Runtime ${stamp}`;
    db.cleanup(
      `DELETE FROM env_bundles WHERE name LIKE 'E2E PodUI %'`
    );
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: AGENT_SLUG,
      name: runtimeName,
      kind: "runtime",
      data: { CLAUDE_LOG_LEVEL: "debug" },
    }) as { id: bigint };
    const runtimeId = runtime.id;

    // Frontend now goes Connect-RPC (binary proto) — capture and decode.
    let capturedLayer: string | undefined;
    await page.route(`**${CREATE_POD_RPC}`, async (route) => {
      if (route.request().method() === "POST") {
        const layer = decodeCreatePodLayer(route.request().postDataBuffer());
        if (typeof layer === "string") capturedLayer = layer;
      }
      await route.continue();
    });

    await terminateAllPods();

    try {
      const worker = new CreateWorkerPage(page, TEST_ORG_SLUG);
      await worker.goto();
      await worker.selectImage(AGENT_SLUG);
      await expect(page.locator("label", { hasText: runtimeName })).toBeVisible();

      const createResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response.url().endsWith(CREATE_POD_RPC),
        { timeout: 20_000 },
      );
      await worker.selectRuntimeBundles([runtimeName]);
      await worker.selectPtyMode();
      await worker.submit();
      const response = await createResponse;
      expect(response.ok()).toBeTruthy();
      await worker.waitForWorkspace();

      const layer = capturedLayer ?? "";
      const useLines = layer
        .split("\n")
        .filter((l) => l.startsWith("USE_ENV_BUNDLE"));
      expect(useLines).toEqual([`USE_ENV_BUNDLE "${runtimeName}"`]);
    } finally {
      if (runtimeId) await cc.envBundle.deleteEnvBundle({ id: runtimeId }).catch(() => null);
      db.cleanup(`DELETE FROM env_bundles WHERE name LIKE 'E2E PodUI %'`);
    }
  });

  test("no-bundle selection omits USE_ENV_BUNDLE from agentfile_layer", async ({
    page,
    api,
    db,
  }) => {
    const cc = await api.connect();
    const { items: runners } = await cc.runner.listAvailableRunners({
      orgSlug: TEST_ORG_SLUG,
    }) as { items: unknown[] };
    expect(runners.length, "dev env must have an online runner").toBeGreaterThan(0);
    const { builtinAgents } = await cc.agent.listAgents({
      orgSlug: TEST_ORG_SLUG,
    }) as { builtinAgents: { slug: string }[] };
    expect(
      builtinAgents?.some((a) => a.slug === AGENT_SLUG),
      "dev env must include the e2e-echo builtin agent",
    ).toBeTruthy();

    // The Pod multi-select auto-checks every primary runtime bundle, so any
    // stray primary for e2e-echo would flip the assertion below.
    // Purge them up-front to keep the empty-selection path testable.
    db.cleanup(
      `DELETE FROM env_bundles WHERE agent_slug = '${AGENT_SLUG}' AND kind = 'runtime' AND kind_primary = TRUE`
    );

    let capturedLayer: string | undefined;
    await page.route(`**${CREATE_POD_RPC}`, async (route) => {
      if (route.request().method() === "POST") {
        const layer = decodeCreatePodLayer(route.request().postDataBuffer());
        // Absent agentfile_layer also counts as "no USE_ENV_BUNDLE".
        capturedLayer = typeof layer === "string" ? layer : "";
      }
      await route.continue();
    });

    await terminateAllPods();

    const worker = new CreateWorkerPage(page, TEST_ORG_SLUG);
    await worker.goto();
    await worker.selectImage(AGENT_SLUG);
    const createResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith(CREATE_POD_RPC),
      { timeout: 20_000 },
    );
    await worker.selectPtyMode();
    await worker.submit();
    const response = await createResponse;
    expect(response.ok()).toBeTruthy();
    await worker.waitForWorkspace();

    expect(capturedLayer ?? "").not.toContain("USE_ENV_BUNDLE");
  });
});
