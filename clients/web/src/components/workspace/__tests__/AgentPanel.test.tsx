import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

const mocks = vi.hoisted(() => ({
  lastRuntimeInput: null as null | Record<string, unknown>,
  podStatus: "initializing",
  sessionEnabled: [] as boolean[],
}));

vi.mock("@/hooks", () => ({
  usePodStatus: () => ({
    podStatus: mocks.podStatus,
    isPodReady: mocks.podStatus === "ready",
    podError: null,
  }),
}));

vi.mock("@/hooks/useAcpRelay", () => ({
  useAcpRelay: vi.fn(),
}));

vi.mock("@/hooks/useAgentSessionLink", () => ({
  useAgentSessionLink: (_podKey: string, enabled: boolean) => {
    mocks.sessionEnabled.push(enabled);
    return enabled
      ? { error: null, loading: false, sessionId: "session-1" }
      : { error: null, loading: false, sessionId: null };
  },
}));

vi.mock("@/hooks/useWorkerControlLease", () => ({
  useWorkerControlLease: () => ({
    acquire: vi.fn(),
    acquiring: false,
    connected: false,
    error: null,
    status: "idle",
  }),
}));

vi.mock("@/lib/wasm-core", () => ({
  getAgentWorkbenchState: () => ({}),
}));

vi.mock("@/stores/pod", () => ({
  usePod: () => ({
    agent: { name: "Pattern Designer" },
    interaction_mode: "acp",
    title: "Pattern preview",
  }),
  usePodStore: (
    selector: (state: { initProgress: Record<string, unknown> }) => unknown,
  ) => selector({ initProgress: {} }),
}));

vi.mock("@/stores/workspace", () => ({
  useWorkspaceStore: (
    selector: (state: {
      panes: unknown[];
      setActivePane: ReturnType<typeof vi.fn>;
      splitPane: ReturnType<typeof vi.fn>;
    }) => unknown,
  ) =>
    selector({
      panes: [],
      setActivePane: vi.fn(),
      splitPane: vi.fn(),
    }),
}));

vi.mock("@do-worker/agent-ui", () => ({
  AgentWorkspace: ({ readOnly }: { readOnly: boolean }) => (
    <div data-readonly={String(readOnly)} data-testid="agent-workspace" />
  ),
  createBuiltinContentRenderers: () => ({}),
}));

vi.mock("../agent-ui/WebAgentWorkbenchRuntime", () => ({
  WebAgentWorkbenchRuntime: class {
    sessionId: string;
    constructor(input: Record<string, unknown>) {
      mocks.lastRuntimeInput = input;
      this.sessionId = String(input.sessionId);
    }
  },
}));

vi.mock("../agent-ui/webAgentWorkbenchArtifactLoader", () => ({
  createWebAgentWorkbenchArtifactLoader: () => vi.fn(),
}));

vi.mock("../AgentPanelHeader", () => ({
  AgentPanelHeader: () => null,
}));

vi.mock("../PodSelectorModal", () => ({
  PodSelectorModal: () => null,
}));

vi.mock("@/components/mobile-worker/WorkerControlOverlay", () => ({
  WorkerControlOverlay: () => <div data-testid="control-overlay" />,
}));

import { AgentPanel } from "../AgentPanel";

describe("AgentPanel artifact access", () => {
  beforeEach(() => {
    mocks.lastRuntimeInput = null;
    mocks.podStatus = "initializing";
    mocks.sessionEnabled = [];
  });

  it("mounts a completed Worker session in read-only mode", () => {
    mocks.podStatus = "completed";

    render(
      <AgentPanel
        paneId="pane-1"
        podKey="pod-1"
        isActive
        showHeader={false}
      />,
    );

    expect(mocks.sessionEnabled).toContain(true);
    expect(mocks.lastRuntimeInput).toMatchObject({ live: false });
    expect(screen.getByTestId("agent-workspace")).toHaveAttribute(
      "data-readonly",
      "true",
    );
    expect(screen.queryByTestId("control-overlay")).not.toBeInTheDocument();
  });

  it("keeps the live controls only for a running Worker", () => {
    mocks.podStatus = "running";

    render(
      <AgentPanel
        paneId="pane-1"
        podKey="pod-1"
        isActive
        showHeader={false}
      />,
    );

    expect(mocks.lastRuntimeInput).toMatchObject({ live: true });
    expect(screen.getByTestId("agent-workspace")).toHaveAttribute(
      "data-readonly",
      "true",
    );
    expect(screen.getByTestId("control-overlay")).toBeInTheDocument();
  });

  it("keeps the loading state before the Worker is readable", () => {
    render(
      <AgentPanel
        paneId="pane-1"
        podKey="pod-1"
        isActive
        showHeader={false}
      />,
    );

    expect(mocks.sessionEnabled).toContain(false);
    expect(screen.queryByTestId("agent-workspace")).not.toBeInTheDocument();
    expect(screen.getByText("Waiting for Pod to be ready...")).toBeInTheDocument();
  });
});
