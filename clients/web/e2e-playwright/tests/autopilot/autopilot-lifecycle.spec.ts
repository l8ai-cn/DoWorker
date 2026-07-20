// Full-chain autopilot decision loop: web → backend → runner → mock control
// agent (claude-stream) → MCP → target pod. decision_type is the upper-cased
// wire enum (CONTINUE / TASK_COMPLETED / NEED_HUMAN_HELP / GIVE_UP).
import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { terminateRegisteredE2EPods } from "../../helpers/pod-cleanup";
import { pollUntil } from "../../helpers/retry";
import {
  createReadyAutopilotTarget,
  createAutopilotForPod,
  collectAutopilotEvents,
  isTerminalStatus,
  hasThinkingDecision,
  hasStatusPhase,
  hasIterationEvent,
  hasIterationWithFiles,
} from "../../helpers/autopilot";
import { getIterationHistory, type AutopilotIterationItem } from "../../helpers/autopilot-actions";

// Each turn holds the pod ~6s to clear the controller's 5s MinTriggerGap, so
// multi-iteration cases need generous time.
const CASE_TIMEOUT = 150_000;

test.describe("Autopilot lifecycle · mock control agent", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  test.afterEach(async () => {
    await terminateRegisteredE2EPods();
  });

  test("completed: decision loop drives the pod then terminates", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);
    const sentinel = `done-${pod.podKey}`;

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 120_000 },
      () =>
        createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: {
            decisions: [
              { type: "continue", reasoning: "step1", send_input: "echo step1\n" },
              { type: "completed", reasoning: sentinel },
            ],
          },
          maxIterations: 5,
        }),
    );

    expect(events.map((e) => e.type)).toContain("autopilot:created");
    // The completed decision flowed end-to-end (mock → claude-stream →
    // DecisionParser → gRPC → eventbus) and carried the injected sentinel.
    expect(
      hasThinkingDecision(events, "COMPLETED", sentinel),
      `thinking: ${JSON.stringify(events.filter((e) => e.type === "autopilot:thinking").map((e) => e.data))}`,
    ).toBe(true);
    expect(hasIterationEvent(events), `events: ${events.map((e) => e.type).join(",")}`).toBe(true);
    expect(hasStatusPhase(events, "completed")).toBe(true);
  });

  test("iteration history persists and is queryable (BUG-1 column)", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);
    let key = "";

    await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 120_000 },
      async () => {
        key = await createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: {
            decisions: [
              { type: "continue", reasoning: "drive", send_input: "echo go\n" },
              { type: "completed", reasoning: "done" },
            ],
          },
          maxIterations: 5,
        });
      },
    );

    // GetIterations reads autopilot_iterations via autopilot_controller_id
    // (migration 000158). A non-empty history proves the write+read path works
    // against the real schema — the regression BUG-1 would surface as empty/500.
    let iters: AutopilotIterationItem[] = [];
    await pollUntil(
      async () => {
        iters = await getIterationHistory(api, key);
        return iters.length >= 1;
      },
      { maxAttempts: 10, intervalMs: 1000, label: `iterations-${key}` },
    );
    expect(iters.some((it) => (it.iterationNumber ?? 0) >= 1)).toBe(true);
  });

  test("give_up: control agent abandons → failed", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 60_000 },
      () =>
        createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "give_up", reasoning: "cannot proceed" }] },
          maxIterations: 5,
        }),
    );

    expect(hasThinkingDecision(events, "GIVE_UP")).toBe(true);
    expect(hasStatusPhase(events, "failed")).toBe(true);
  });

  test("max_iterations: cap reached → terminates", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    // maxIterations:1 → iteration 1 drives the pod, the next waiting edge hits
    // the cap. Every decision is continue+send_input so the pod keeps cycling.
    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 90_000 },
      () =>
        createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "continue", reasoning: "loop", send_input: "echo loop\n" }] },
          maxIterations: 1,
        }),
    );

    expect(hasStatusPhase(events, "max_iterations")).toBe(true);
  });

  test("progress tracker reports changed files in iteration events", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    // autopilot_fs git-inits the sandbox + drops a probe file, so the runner's
    // ProgressTracker git-diff snapshot surfaces files_changed on the wire.
    const pod = await createReadyAutopilotTarget(api, { scenario: "autopilot_fs" });

    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 120_000 },
      () =>
        createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: {
            decisions: [
              { type: "continue", reasoning: "step", send_input: "echo go\n" },
              { type: "completed", reasoning: "done" },
            ],
          },
          maxIterations: 5,
        }),
    );

    expect(
      hasIterationWithFiles(events),
      `iteration files: ${JSON.stringify(events.filter((e) => e.type === "autopilot:iteration").map((e) => e.data.files_changed))}`,
    ).toBe(true);
  });

  test("circuit breaker: consecutive control errors → failed", async ({ api }) => {
    test.setTimeout(CASE_TIMEOUT);
    const { token } = await api.login();
    const pod = await createReadyAutopilotTarget(api);

    // Three "error" decisions make the control agent emit empty output; the
    // controller retries (pod stays waiting) until MaxConsecutiveErrors=3 trips.
    const events = await collectAutopilotEvents(
      { token, until: isTerminalStatus, timeoutMs: 90_000 },
      () =>
        createAutopilotForPod(api, {
          targetPodKey: pod.podKey,
          script: { decisions: [{ type: "error" }, { type: "error" }, { type: "error" }] },
          maxIterations: 5,
        }),
    );

    // phase=failed alone matches give_up too; the error-phase iteration event
    // proves it was the consecutive-error circuit breaker.
    expect(hasIterationEvent(events, "error"), `events: ${events.map((e) => e.type).join(",")}`).toBe(true);
    expect(hasStatusPhase(events, "failed")).toBe(true);
  });
});
