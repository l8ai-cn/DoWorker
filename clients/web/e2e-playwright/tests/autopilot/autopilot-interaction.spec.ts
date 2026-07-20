// Autopilot user-interaction + control-action coverage (the review's state-machine
// blind spots). Same fully-mocked chain as autopilot-lifecycle.spec.ts; here the
// spec drives the control RPCs (approve / pause / resume / takeover / handback /
// stop) and asserts the resulting phase transitions on the wire.
import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import {
  createReadyAutopilotTarget,
  createAutopilotForPod,
  collectAutopilotEvents,
  isTerminalStatus,
  hasThinkingDecision,
  hasStatusPhase,
} from "../../helpers/autopilot";
import {
  pauseAutopilot,
  resumeAutopilot,
  stopAutopilot,
  takeoverAutopilot,
  handbackAutopilot,
  approveAutopilot,
} from "../../helpers/autopilot-actions";

const CASE_TIMEOUT = 150_000;
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
const stoppedPhase = (t: string, d: Record<string, unknown>) =>
  t === "autopilot:status_changed" && d.phase === "stopped";

test.describe("Autopilot interaction · mock control agent", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  test.afterEach(async () => {
    await terminateRegisteredE2EPods();
  });

  test("need_human_help: pauses for approval, approve resumes to completed", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);
    const sentinel = `done-${pod.podKey}`;

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 120_000 },
      async () => {
        const key = await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: {
            decisions: [
              { type: "need_help", reasoning: "need a human" },
              { type: "completed", reasoning: sentinel },
            ],
          },
          maxIterations: 5,
        });
        // Iteration 1 → need_help → waiting_approval. Wait past the controller's
        // 5s MinTriggerGap so the post-approve OnPodWaiting isn't deduplicated.
        await sleep(8000);
        await approveAutopilot(api, key, { continueExecution: true, additionalIterations: 3 });
      },
    );

    expect(hasThinkingDecision(events, "NEED_HUMAN_HELP")).toBe(true);
    expect(hasStatusPhase(events, "waiting_approval")).toBe(true);
    // Approval resumed the loop and the next decision completed it.
    expect(hasThinkingDecision(events, "COMPLETED", sentinel)).toBe(true);
    expect(hasStatusPhase(events, "completed")).toBe(true);
  });

  test("need_human_help: deny approval stops the controller", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 60_000 },
      async () => {
        const key = await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "need_help", reasoning: "need a human" }] },
          maxIterations: 5,
        });
        await sleep(4000);
        await approveAutopilot(api, key, { continueExecution: false });
      },
    );

    expect(hasStatusPhase(events, "waiting_approval")).toBe(true);
    expect(hasStatusPhase(events, "stopped")).toBe(true);
  });

  test("pause → resume → stop", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    const events = await collectAutopilotEvents(
      { token, until: stoppedPhase, timeoutMs: 60_000 },
      async () => {
        const key = await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "continue", reasoning: "loop", send_input: "echo loop\n" }] },
          maxIterations: 10,
        });
        await sleep(2000);
        await pauseAutopilot(api, key);
        await sleep(2500);
        await resumeAutopilot(api, key);
        await sleep(2500);
        await stopAutopilot(api, key);
      },
    );

    expect(hasStatusPhase(events, "paused")).toBe(true);
    expect(hasStatusPhase(events, "stopped")).toBe(true);
  });

  test("takeover → handback → stop", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    const events = await collectAutopilotEvents(
      { token, until: stoppedPhase, timeoutMs: 60_000 },
      async () => {
        const key = await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "continue", reasoning: "loop", send_input: "echo loop\n" }] },
          maxIterations: 10,
        });
        await sleep(2000);
        await takeoverAutopilot(api, key);
        await sleep(2500);
        await handbackAutopilot(api, key);
        await sleep(2500);
        await stopAutopilot(api, key);
      },
    );

    expect(hasStatusPhase(events, "user_takeover")).toBe(true);
    expect(hasStatusPhase(events, "stopped")).toBe(true);
  });
});
