import type { ApiFixture } from "../fixtures/api.fixture";
import { createE2EEchoPod } from "./e2e-worker-spec";
import { TEST_ORG_SLUG, getApiBaseUrl } from "./env";
import { unregisterE2ECreatedPod } from "./pod-cleanup";
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

interface Pod { podKey: string; runnerId: bigint }

// createMockAgentPod spawns a pod backed by the e2e-mock-agent binary via
// the same WorkerSpec contract used by production Pod creation.
// The returned `cleanup` must be invoked from afterEach to avoid quota bleed.
export async function createMockAgentPod(
  api: ApiFixture,
  opts: CreateMockPodOptions,
): Promise<MockAgentPod> {
  const cc = await api.connect();
  const resp = await createE2EEchoPod(cc, {
    mode: opts.mode,
    scenario: opts.scenario,
    prompt: opts.prompt,
    alias: opts.alias,
  }) as { pod?: Pod };
  const podKey = resp.pod?.podKey;
  const runnerId = resp.pod?.runnerId;
  if (!podKey || !runnerId) {
    throw new Error("createMockAgentPod returned incomplete pod placement");
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
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
      unregisterE2ECreatedPod(podKey);
    },
  };
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
