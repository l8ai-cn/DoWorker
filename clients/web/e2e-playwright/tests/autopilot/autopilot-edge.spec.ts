// Autopilot edge cases: concurrent controllers stay isolated, and a target pod
// dying mid-run tears its controller down (runner ac.Stop → terminated →
// backend status_changed phase=stopped).
import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import {
  createReadyAutopilotTarget,
  createAutopilotForPod,
  collectAutopilotEvents,
  isTerminalStatus,
  hasStatusPhase,
} from "../../helpers/autopilot";

const CASE_TIMEOUT = 180_000;

const completedScript = (sentinel: string) => ({
  decisions: [
    { type: "continue" as const, reasoning: "step", send_input: "echo go\n" },
    { type: "completed" as const, reasoning: sentinel },
  ],
});

test.describe("Autopilot edge cases · mock control agent", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  test.afterEach(async () => {
    await terminateRegisteredE2EPods();
  });

  test("concurrent controllers complete independently", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const [podA, podB] = await Promise.all([
      createReadyAutopilotTarget(api),
      createReadyAutopilotTarget(api),
    ]);

    const completed = new Set<string>();
    await collectAutopilotEvents(
      {
        token,
        timeoutMs: 150_000,
        until: (type, data) => {
          if (type === "autopilot:status_changed" && data.phase === "completed") {
            completed.add(String(data.autopilot_controller_key));
          }
          return completed.size >= 2;
        },
      },
      async () => {
        await Promise.all([
          createAutopilotForPod(api, { targetPodKey: podA.podKey, script: completedScript(`a-${podA.podKey}`) }),
          createAutopilotForPod(api, { targetPodKey: podB.podKey, script: completedScript(`b-${podB.podKey}`) }),
        ]);
      },
    );

    expect(completed.size, `distinct completed controllers: ${[...completed].join(", ")}`).toBe(2);
  });

  test("target pod terminated mid-run tears down the controller", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 90_000 },
      async () => {
        await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "continue", reasoning: "loop", send_input: "echo loop\n" }] },
          maxIterations: 20,
        });
        // Let iteration 1 run, then kill the pod out from under the controller.
        await new Promise((r) => setTimeout(r, 9000));
        await pod.cleanup();
      },
    );

    expect(hasStatusPhase(events, "stopped")).toBe(true);
  });
});
