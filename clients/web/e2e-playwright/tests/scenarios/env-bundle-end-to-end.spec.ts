import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { uniqueSuffix } from "../../helpers/test-data";
import { SettingsNavPage } from "../../pages/settings/settings-nav.page";
import {
  createPodAndWaitRunning,
  readEnvDumpFromRunner,
  clearRunnerDumps,
} from "../../helpers/env-bundle-e2e";

const KIND_RUNTIME = "runtime";

/**
 * EnvBundle end-to-end causal chain:
 *
 *   Settings UI → bundle row → Pod create dialog → backend agentfile eval →
 *   runner gRPC CreatePodCommand → bash spawn with env → e2e-echo writes
 *   env dump to /tmp/e2e-echo-env-dump-<pid>
 *
 * We verify the dump file via `docker exec cat` inside the runner
 * container. This proves the full Settings-UI → child-process env path
 * without depending on PTY streaming (which is async / unreliable for
 * daemon-managed pods).
 *
 * Uses the e2e-echo builtin agent, modified by migration 000150 to write
 * whitelisted env vars to a sandbox file on startup.
 */
const AGENT_SLUG = "e2e-echo";
const BUNDLE_PREFIX = "e2e-bundle-chain";

const unique = (label: string) => `${BUNDLE_PREFIX}-${label}-${uniqueSuffix()}`;

/**
 * Drive Settings UI to create a runtime EnvBundle.
 */
async function createBundleViaSettingsUI(args: {
  page: import("@playwright/test").Page;
  kind: typeof KIND_RUNTIME;
  name: string;
  envKey: string;
  envValue: string;
}): Promise<void> {
  const { page, name, envKey, envValue } = args;
  const selectors = {
    openDialog: async () => {
      const heading = page
        .getByRole("heading", { name: /Runtime Env Variables|运行时环境变量/i })
        .first();
      await heading.waitFor({ state: "visible", timeout: 10_000 });
      await heading.scrollIntoViewIfNeeded();
      await heading
        .locator(
          'xpath=following::button[normalize-space(.)="Add" or normalize-space(.)="添加"][1]',
        )
        .click();
    },
    nameInput: "#runtime-name",
    fillEnv: async () => {
      await page
        .locator('input[placeholder="ENV_NAME"], input[placeholder="环境变量名"]')
        .first()
        .fill(envKey);
      await page
        .locator('input[placeholder="Value"], input[placeholder="值"]')
        .first()
        .fill(envValue);
    },
  };

  const nav = new SettingsNavPage(page, TEST_ORG_SLUG);
  await nav.goto("personal", `agents/${AGENT_SLUG}`);
  await selectors.openDialog();
  await page.locator(selectors.nameInput).waitFor({ state: "visible", timeout: 5000 });
  await page.locator(selectors.nameInput).fill(name);
  await selectors.fillEnv();
  await page.getByRole("button", { name: /^(创建|Create)$/ }).click();
  await page.locator(selectors.nameInput).waitFor({ state: "hidden", timeout: 5000 });
  await page.getByText(name, { exact: false }).first().waitFor({ timeout: 5000 });
}

test.describe("EnvBundle end-to-end (Settings UI → Pod → child env)", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
    await terminateAllPods();
  });

  test.afterEach(async ({ db }) => {
    // Only terminateAllPods is async; the other two cleanups are sync and
    // gain nothing from Promise.all wrapping. Order: pods first (drops
    // dump files via their own teardown), then SQL + dump rm.
    await terminateAllPods();
    db.cleanup(`DELETE FROM env_bundles WHERE name LIKE '${BUNDLE_PREFIX}%'`);
    clearRunnerDumps();
  });

  test("runtime bundle: Settings UI → Pod create → env injected to child process", async ({
    page,
    api,
  }) => {
    const bundleName = unique("rt");
    const envKey = "E2E_TEST_BUNDLE_RUNTIME";
    const envValue = `runtime-marker-${Date.now()}`;

    await createBundleViaSettingsUI({
      page,
      kind: KIND_RUNTIME,
      name: bundleName,
      envKey,
      envValue,
    });

    await createPodAndWaitRunning({
      page,
      api,
      agentSlug: AGENT_SLUG,
      selectRuntimeBundleNames: [bundleName],
    });

    // The agent's env dump is written by the spawned child, which can lag
    // behind the pod reaching "running" (createPodAndWaitRunning only gates on
    // pod status, not on the child having flushed its env). Poll the dump
    // until the injected var lands rather than reading once and racing it.
    await expect(async () => {
      const dump = await readEnvDumpFromRunner();
      expect(dump).toContain(`${envKey}=${envValue}`);
    }).toPass({ timeout: 20_000 });
  });
});
