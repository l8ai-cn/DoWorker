import { test, expect } from "../../fixtures/index";
import { fromBinary } from "@bufbuild/protobuf";
import { CreatePodRequestSchema } from "../../../../../proto/gen/ts/pod/v1/pod_pb";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { clearAuthRateLimit } from "../../helpers/redis";
import { CreatePodModal } from "../../pages/modals/create-pod.modal";

/**
 * Pod create × EnvBundle: step-1 credential Select + runtime checkboxes.
 * Wire contract checked via CreatePod request interception (agentfile_layer).
 */
const CREATE_POD_RPC = "/proto.pod.v1.PodService/CreatePod";

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

  test("Pod create attaches credential first then runtime in order", async ({
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
    const claudeCode = builtinAgents?.find((a) => a.slug === "claude-code");
    expect(claudeCode, "dev env must include the claude-code builtin agent").toBeTruthy();

    const stamp = Date.now();
    const credName = `E2E PodUI Cred ${stamp}`;
    const runtimeName = `E2E PodUI Runtime ${stamp}`;
    db.cleanup(
      `DELETE FROM env_bundles WHERE name LIKE 'E2E PodUI %'`
    );
    const cred = await cc.envBundle.createEnvBundle({
      agentSlug: "claude-code",
      name: credName,
      kind: "credential",
      data: { ANTHROPIC_API_KEY: "sk-ant-e2e-multi" },
    }) as { id: bigint };
    const credId = cred.id;
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: "claude-code",
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
      await page.goto(`/${TEST_ORG_SLUG}/workspace`);
      await page.waitForLoadState("load");

      const newPodBtn = page
        .getByRole("button", { name: /new pod|create new pod|新建 pod|新建 worker|新建环境/i })
        .first();
      await newPodBtn.waitFor({ state: "visible", timeout: 15_000 });
      await newPodBtn.click();
      await page.waitForURL(new RegExp(`/${TEST_ORG_SLUG}/workers/new`), {
        timeout: 15_000,
      });

      const modal = new CreatePodModal(page);
      await modal.waitForOpen();
      await modal.selectAgent("claude-code");

      // Credential + runtime pickers appear on step 1 after image select.
      await expect(modal.credentialTrigger()).toBeVisible({ timeout: 10_000 });
      await modal.credentialTrigger().click();
      await expect(modal.credentialOption(credName)).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(modal.runtimeBundleLabel(runtimeName)).toBeVisible();

      await modal.selectCredential(credName);
      await modal.selectRuntimeBundles([runtimeName]);

      await modal.submit();
      await modal.waitForClosed(15_000);

      // Two USE_ENV_BUNDLE lines: credential first, runtime after.
      const layer = capturedLayer ?? "";
      const useLines = layer
        .split("\n")
        .filter((l) => l.startsWith("USE_ENV_BUNDLE"));
      expect(useLines).toEqual([
        `USE_ENV_BUNDLE "${credName}"`,
        `USE_ENV_BUNDLE "${runtimeName}"`,
      ]);
    } finally {
      if (credId) await cc.envBundle.deleteEnvBundle({ id: credId }).catch(() => null);
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
    const claudeCode = builtinAgents?.find((a) => a.slug === "claude-code");
    expect(claudeCode, "dev env must include the claude-code builtin agent").toBeTruthy();

    // The Pod multi-select auto-checks every primary bundle, so any
    // stray primary for claude-code would flip the assertion below.
    // Purge them up-front to keep the empty-selection path testable.
    db.cleanup(
      `DELETE FROM env_bundles WHERE agent_slug = 'claude-code' AND kind_primary = TRUE`
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

    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");

    const newPodBtn = page
      .getByRole("button", { name: /new pod|create new pod|新建 pod|新建 worker|新建环境/i })
      .first();
    await newPodBtn.waitFor({ state: "visible", timeout: 15_000 });
    await newPodBtn.click();
    await page.waitForURL(new RegExp(`/${TEST_ORG_SLUG}/workers/new`), {
      timeout: 15_000,
    });

    const modal = new CreatePodModal(page);
    await modal.waitForOpen();
    await modal.selectAgent("claude-code");
    // Leave the default empty selection alone (we purged primaries above).
    await modal.submit();
    await modal.waitForClosed(15_000);

    expect(capturedLayer ?? "").not.toContain("USE_ENV_BUNDLE");
  });
});
