import type { ApiFixture } from "../fixtures/api.fixture";
import { TEST_ORG_SLUG } from "./env";
import { subscribeEvents } from "./eventbus-stream";
import { createMockAgentPod, type MockAgentPod } from "./mock-agent";
import { pollUntil } from "./retry";

export interface ControlDecision {
  type: "continue" | "completed" | "need_help" | "give_up";
  reasoning?: string;
  send_input?: string;
}

export interface AutopilotScript {
  observe?: boolean;
  decisions: ControlDecision[];
}

export interface CollectedEvent {
  type: string;
  data: Record<string, unknown>;
}

const SETTLE_MS = 500;

// AutopilotController.Start() only fires the first iteration when the pod is
// "waiting" at attach time; gate here to remove the race.
export async function waitForPodWaiting(
  api: ApiFixture,
  podKey: string,
  opts: { maxAttempts?: number } = {},
): Promise<void> {
  const cc = await api.connect();
  await pollUntil(
    async () => {
      const { items } = (await cc.pod.listPods({ orgSlug: TEST_ORG_SLUG })) as {
        items?: Array<{ podKey?: string; agentStatus?: string }>;
      };
      const mine = items?.find((p) => p.podKey === podKey);
      return mine?.agentStatus === "waiting";
    },
    // Mock autopilot scenario sleeps 6s before printing the waiting prompt;
    // CI runner/gRPC lag needs headroom beyond that.
    { maxAttempts: opts.maxAttempts ?? 45, intervalMs: 1000, label: `pod-${podKey}-waiting` },
  );
}

// Retries on a fresh pod: a dev-env flake leaves the mock producing no PTY
// output on some runner instances, so the pod never reaches "waiting".
export async function createReadyAutopilotTarget(
  api: ApiFixture,
  opts: { attempts?: number; scenario?: "autopilot" | "autopilot_fs" } = {},
): Promise<MockAgentPod> {
  const attempts = opts.attempts ?? 3;
  const scenario = opts.scenario ?? "autopilot";
  let lastErr: unknown;
  for (let i = 0; i < attempts; i++) {
    const pod = await createMockAgentPod(api, { mode: "pty", scenario });
    try {
      await waitForPodWaiting(api, pod.podKey, { maxAttempts: 45 });
      return pod;
    } catch (e) {
      lastErr = e;
      await pod.cleanup();
    }
  }
  throw new Error(`createReadyAutopilotTarget: no pod reached waiting in ${attempts} attempts: ${lastErr}`);
}

// Attaches a mock control agent (claude-stream); `script` is injected via
// control_prompt_template, replayed one decision per iteration over MCP.
export async function createAutopilotForPod(
  api: ApiFixture,
  opts: { targetPodKey: string; script: AutopilotScript; maxIterations?: number },
): Promise<string> {
  const cc = await api.connect();
  const ctrl = (await cc.autopilot.createAutopilotController({
    orgSlug: TEST_ORG_SLUG,
    podKey: opts.targetPodKey,
    prompt: "drive the target pod to completion",
    controlAgentSlug: "e2e-mock-agent",
    controlPromptTemplate: JSON.stringify({
      observe: opts.script.observe ?? true,
      decisions: opts.script.decisions,
    }),
    maxIterations: opts.maxIterations ?? 10,
  })) as { autopilotControllerKey?: string };

  if (!ctrl.autopilotControllerKey) {
    throw new Error(`createAutopilotForPod: missing key in ${JSON.stringify(ctrl)}`);
  }
  return ctrl.autopilotControllerKey;
}

// collectAutopilotEvents subscribes to the EventBus, runs `action`, and
// collects every event until `until` matches (e.g. a terminal status) or the
// timeout elapses. Returns all events in arrival order so specs can assert on
// the whole decision-loop sequence, not just one event.
export async function collectAutopilotEvents(
  opts: {
    token: string;
    until: (type: string, data: Record<string, unknown>) => boolean;
    timeoutMs?: number;
    onEvent?: (type: string, data: Record<string, unknown>) => Promise<void> | void;
  },
  action: () => Promise<void>,
): Promise<CollectedEvent[]> {
  const timeout = opts.timeoutMs ?? 30_000;
  const ctrl = new AbortController();
  const events: CollectedEvent[] = [];

  const drain = (async () => {
    try {
      for await (const ev of subscribeEvents({ token: opts.token, orgSlug: TEST_ORG_SLUG, signal: ctrl.signal })) {
        let data: Record<string, unknown>;
        try {
          data = JSON.parse(ev.dataJson) as Record<string, unknown>;
        } catch {
          data = {};
        }
        events.push({ type: ev.type, data });
        if (opts.onEvent) await opts.onEvent(ev.type, data);
        if (opts.until(ev.type, data)) {
          ctrl.abort();
          return;
        }
      }
    } catch {
      /* abort or clean close */
    }
  })();

  await new Promise((r) => setTimeout(r, SETTLE_MS));
  await action();

  await Promise.race([drain, new Promise((r) => setTimeout(r, timeout))]);
  ctrl.abort();
  return events;
}

export function isTerminalStatus(type: string, data: Record<string, unknown>): boolean {
  return (
    type === "autopilot:status_changed" &&
    ["completed", "failed", "stopped", "max_iterations"].includes(String(data.phase))
  );
}

export function statusEventsWithPhase(events: CollectedEvent[], phase: string): CollectedEvent[] {
  return events.filter((e) => e.type === "autopilot:status_changed" && e.data.phase === phase);
}

// hasThinkingDecision asserts a thinking event carried a decision whose
// upper-cased decision_type contains `typeFragment` (the wire enum, e.g.
// "TASK_COMPLETED" for completed, "GIVE_UP", "NEED_HUMAN_HELP", "CONTINUE")
// and, when given, the exact reasoning string.
export function hasThinkingDecision(
  events: CollectedEvent[],
  typeFragment: string,
  reasoning?: string,
): boolean {
  return events.some(
    (e) =>
      e.type === "autopilot:thinking" &&
      String(e.data.decision_type).toUpperCase().includes(typeFragment.toUpperCase()) &&
      (reasoning === undefined || e.data.reasoning === reasoning),
  );
}

export function hasStatusPhase(events: CollectedEvent[], phase: string): boolean {
  return statusEventsWithPhase(events, phase).length > 0;
}

// hasIterationEvent matches an autopilot:iteration event, optionally with a
// specific phase ("started" / "action_sent" / "error" / "completed" / …).
export function hasIterationEvent(events: CollectedEvent[], phase?: string): boolean {
  return events.some(
    (e) =>
      e.type === "autopilot:iteration" &&
      Number(e.data.iteration) >= 1 &&
      (phase === undefined || e.data.phase === phase),
  );
}

// hasIterationWithFiles matches an iteration event whose files_changed (from the
// runner ProgressTracker git-diff snapshot) is non-empty.
export function hasIterationWithFiles(events: CollectedEvent[]): boolean {
  return events.some(
    (e) =>
      e.type === "autopilot:iteration" &&
      Array.isArray(e.data.files_changed) &&
      (e.data.files_changed as unknown[]).length > 0,
  );
}
