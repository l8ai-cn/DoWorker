import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

const mocks = vi.hoisted(() => ({
  lastRuntimeInput: null as null | Record<string, unknown>,
  presentation: "",
  podStatus: "initializing",
  sessionEnabled: [] as boolean[],
  workspaceArtifacts: [{
    artifactId: "workspace:output/final.mp4",
    filename: "final.mp4",
  }],
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

vi.mock("@agent-cloud/agent-ui", () => ({
  AgentWorkspace: ({
    presentation,
    readOnly,
    workspaceArtifacts,
  }: {
    presentation: string;
    readOnly: boolean;
    workspaceArtifacts: unknown[];
  }) => {
    mocks.presentation = presentation;
    return (
      <div
        data-readonly={String(readOnly)}
        data-testid="agent-workspace"
        data-workspace-artifacts={String(workspaceArtifacts.length)}
      />
    );
  },
  createBuiltinContentRenderers: () => ({}),
  createBuiltinToolRenderers: () => ({}),
}));

vi.mock("../agent-ui/useAgentPanelRuntime", () => ({
  useAgentPanelRuntime: (input: Record<string, unknown>) => {
    mocks.lastRuntimeInput = input;
    return input.sessionId ? { sessionId: input.sessionId } : null;
  },
}));

vi.mock("../agent-ui/usePodWorkspaceArtifacts", () => ({
  usePodWorkspaceArtifacts: () => ({
    artifacts: mocks.workspaceArtifacts,
    error: null,
  }),
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
    mocks.presentation = "";
    mocks.podStatus = "initializing";
    mocks.sessionEnabled = [];
  });

  it.each(["completed", "orphaned"])(
    "mounts a %s Worker session in read-only mode",
    (podStatus) => {
      mocks.podStatus = podStatus;

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
      expect(screen.getByTestId("agent-workspace")).toHaveAttribute(
        "data-workspace-artifacts",
        "1",
      );
      expect(mocks.presentation).toBe("developer");
      expect(screen.queryByTestId("control-overlay")).not.toBeInTheDocument();
    },
  );

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
    expect(mocks.presentation).toBe("developer");
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
