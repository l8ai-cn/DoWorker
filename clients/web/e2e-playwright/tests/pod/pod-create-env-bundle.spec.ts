import { test, expect } from "../../fixtures/index";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { clearAuthRateLimit } from "../../helpers/redis";
import {
  clearRunnerDumps,
  createPodAndWaitRunning,
  readEnvDumpFromRunner,
} from "../../helpers/env-bundle-e2e";
import { uniqueSuffix } from "../../helpers/test-data";

/**
 * EnvBundle selection belongs to AgentFile, and e2e-echo is an internal
 * test-only agent. These tests exercise that supported typed path through the
 * Runner instead of trying to select e2e-echo in the public WorkerSpec form.
 */
const AGENT_SLUG = "e2e-echo";

test.describe("Pod create — EnvBundle binding", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
    clearRunnerDumps();
    await terminateAllPods();
  });

  test.afterEach(async () => {
    await terminateAllPods();
    clearRunnerDumps();
  });

  test("selected runtime bundle reaches the Runner child environment", async ({ api }) => {
    const cc = await api.connect();
    const runtimeName = `e2e-pod-runtime-${uniqueSuffix()}`;
    const envKey = `E2E_POD_RUNTIME_${Date.now()}`;
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
    const envKey = `E2E_POD_UNSELECTED_${Date.now()}`;
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
