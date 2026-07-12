import type { ApiFixture } from "../fixtures/api.fixture";
import { E2E_ECHO_AGENT_SLUG, pickE2EEchoRunner } from "./e2e-echo-runner";
import { TEST_ORG_SLUG, getApiBaseUrl } from "./env";
import { pollUntil } from "./retry";

export type MockAgentMode = "pty" | "acp";

// Scenario names registered in //runner/internal/agents/mockagent/scenarios.go.
// Keep in sync with that file and backend/migrations/000151..000153.
export type MockAgentScenario =
  | "echo"
  | "autopilot"
  | "autopilot_fs"
  | "streaming_3"
  | "thinking_then_answer"
  | "tool_call_edit"
  | "permission_request_edit"
  | "config_change_plan"
  | "fail_after_1s"
  | "malformed_json"
  | "tool_call_failed"
  | "log_warnings"
  | "loopal_panels"
  | "permission_modes_loopal";

export interface CreateMockPodOptions {
  mode: MockAgentMode;
  scenario?: MockAgentScenario;
  prompt?: string;
  alias?: string;
  waitForRelay?: boolean;
}

export interface MockAgentPod {
  podKey: string;
  runnerId: bigint;
  cleanup: () => Promise<void>;
}

interface Runner { id: bigint; nodeId?: string }
interface Pod { podKey: string }

// createMockAgentPod spawns a pod backed by the e2e-mock-agent binary via
// Connect-RPC (PodService.CreatePod). Throws when no runner is online —
// the e2e suite contract is "dev env has at least one online runner", so
// returning null here would silently mask a missing prerequisite.
// The returned `cleanup` must be invoked from afterEach to avoid quota bleed.
export async function createMockAgentPod(
  api: ApiFixture,
  opts: CreateMockPodOptions,
): Promise<MockAgentPod> {
  const cc = await api.connect();
  const { items: runners } = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG }) as { items?: Runner[] };
  const runnerId = pickE2EEchoRunner(runners).id;

  const input: Record<string, unknown> = {
    orgSlug: TEST_ORG_SLUG,
    runnerId,
    agentSlug: E2E_ECHO_AGENT_SLUG,
    agentfileLayer: buildAgentfileLayer(opts),
  };
  if (opts.alias) input.alias = opts.alias;

  const resp = await cc.pod.createPod(input) as { pod?: Pod };
  const podKey = resp.pod?.podKey;
  if (!podKey) {
    throw new Error(`createMockAgentPod missing podKey: ${JSON.stringify(resp)}`);
  }

  if (opts.waitForRelay !== false) {
    // CreatePod confirms dispatch, not that the runner has registered the Pod.
    // Gate navigation on the browser's actual Relay connection contract.
    await pollUntil(
      async () => {
        try {
          const connection = await cc.pod.getPodConnection({
            orgSlug: TEST_ORG_SLUG,
            podKey,
          }) as { relayUrl?: string };
          return Boolean(connection.relayUrl);
        } catch {
          return false;
        }
      },
      { maxAttempts: 30, intervalMs: 1000, label: `pod-${podKey}-relay-ready` },
    );
  }

  return {
    podKey,
    runnerId,
    cleanup: async () => {
      try {
        await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
      } catch {
        // best-effort: tests should not fail because cleanup raced
      }
    },
  };
}

function buildAgentfileLayer(opts: CreateMockPodOptions): string {
  // Autonomous policy can force ACP unless the layer declares a mode.
  const lines: string[] = [`MODE ${opts.mode}`];
  if (opts.scenario && opts.scenario !== "echo") {
    lines.push(`CONFIG scenario = "${opts.scenario}"`);
  }
  // PROMPT travels through the AgentFile layer in the Connect-RPC create
  // path — CreatePodRequest does not expose a top-level prompt field.
  if (opts.prompt) {
    // Escape backslashes and double-quotes for the AgentFile single-line
    // string syntax (`PROMPT "..."`). Tests pass plain ASCII so this is
    // sufficient — no multi-line / unicode-escape handling needed.
    const safe = opts.prompt.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
    lines.push(`PROMPT "${safe}"`);
  }
  return lines.length > 0 ? lines.join("\n") + "\n" : "";
}

// Returns the workspace URL for a given pod, which renders the
// AcpActivityStream / AcpPromptInput / AcpPermissionDialog stack.
// Pod selection travels via the `pod` query param — the workspace page
// reads it through useSearchParams and calls addPane(podKey) once the
// store hydrates.
export function workspaceUrlForPod(podKey: string): string {
  return `/${TEST_ORG_SLUG}/workspace?pod=${encodeURIComponent(podKey)}`;
}

// Returns the standalone Loopal control-console URL for a pod, which renders
// the bg-shell / cron / task / topology / mcp / goal panels and control bars
// (distinct from the workspace route — this is /[org]/loopal/[podKey]).
export function loopalConsoleUrlForPod(podKey: string): string {
  return `/${TEST_ORG_SLUG}/loopal/${encodeURIComponent(podKey)}`;
}

// getApiBaseUrl re-export for tests that need it without an extra import.
export { getApiBaseUrl };
