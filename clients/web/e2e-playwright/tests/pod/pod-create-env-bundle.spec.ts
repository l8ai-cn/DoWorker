import { test, expect } from "../../fixtures/index";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  clearRunnerDumps,
  createPodAndWaitRunning,
  readEnvDumpFromRunner,
} from "../../helpers/env-bundle-e2e";
import { uniqueSuffix } from "../../helpers/test-data";

// e2e-echo is selectable only in the E2E environment. These tests verify that
// EnvBundle references survive the same WorkerSpec resource path used by pods.
const AGENT_SLUG = "e2e-echo";

test.describe("Pod create — EnvBundle binding", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
    clearRunnerDumps();
    await terminateRegisteredE2EPods();
  });

  test.afterEach(async () => {
    await terminateRegisteredE2EPods();
    clearRunnerDumps();
  });

  test("selected runtime bundle reaches the Runner child environment", async ({ api }) => {
    const cc = await api.connect();
    const runtimeName = `e2e-pod-runtime-${uniqueSuffix()}`;
    const envKey = `E2E_TEST_POD_RUNTIME_${Date.now()}`;
    const envValue = `runtime-marker-${Date.now()}`;
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: AGENT_SLUG,
      name: runtimeName,
      kind: "runtime",
      data: { [envKey]: envValue },
    }) as { id: bigint };

    try {
      await createPodAndWaitRunning({
        api,
        agentSlug: AGENT_SLUG,
        runtimeBundleNames: [runtimeName],
      });
      await expect(async () => {
        expect(await readEnvDumpFromRunner()).toContain(`${envKey}=${envValue}`);
      }).toPass({ timeout: 20_000 });
    } finally {
      await cc.envBundle.deleteEnvBundle({ id: runtime.id }).catch(() => null);
    }
  });

  test("unselected runtime bundle is not injected into the child environment", async ({ api }) => {
    const cc = await api.connect();
    const runtimeName = `e2e-pod-unselected-${uniqueSuffix()}`;
    const envKey = `E2E_TEST_POD_UNSELECTED_${Date.now()}`;
    const runtime = await cc.envBundle.createEnvBundle({
      agentSlug: AGENT_SLUG,
      name: runtimeName,
      kind: "runtime",
      data: { [envKey]: "must-not-be-present" },
    }) as { id: bigint };

    try {
      await createPodAndWaitRunning({ api, agentSlug: AGENT_SLUG });
      await expect(async () => {
        expect(await readEnvDumpFromRunner()).not.toContain(`${envKey}=`);
      }).toPass({ timeout: 20_000 });
    } finally {
      await cc.envBundle.deleteEnvBundle({ id: runtime.id }).catch(() => null);
    }
  });
});
