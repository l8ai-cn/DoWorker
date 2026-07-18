import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

const mocks = vi.hoisted(() => ({
  podStatus: "initializing",
}));

vi.mock("@/hooks", () => ({
  usePodStatus: () => ({
    podStatus: mocks.podStatus,
    isPodReady: mocks.podStatus === "ready",
    podError: null,
  }),
}));

vi.mock("@/hooks/useMigratedSessionHydration", () => ({
  useMigratedSessionHydration: vi.fn(),
}));

vi.mock("@/stores/pod", () => ({
  usePod: () => ({
    agent: { name: "Pattern Designer" },
    title: "Pattern preview",
  }),
  usePodStore: (
    selector: (state: { initProgress: Record<string, unknown> }) => unknown,
  ) => selector({ initProgress: {} }),
}));

vi.mock("@/stores/workspace", () => ({
  useWorkspaceStore: (
    selector: (state: {
      setActivePane: ReturnType<typeof vi.fn>;
      splitPane: ReturnType<typeof vi.fn>;
      panes: unknown[];
    }) => unknown,
  ) =>
    selector({
      setActivePane: vi.fn(),
      splitPane: vi.fn(),
      panes: [],
    }),
}));

vi.mock("@do-worker/agent-ui", () => ({
  AgentWorkspace: () => <div data-testid="agent-workspace" />,
}));

vi.mock("../agent-ui/WebAcpSessionRuntime", () => ({
  WebAcpSessionRuntime: class {
    sessionId = "web-acp:pod-1";
  },
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
    mocks.podStatus = "initializing";
  });

  it.each(["completed", "orphaned"])(
    "keeps the artifact workspace mounted for a %s Worker",
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

      expect(screen.getByTestId("agent-workspace")).toBeInTheDocument();
      expect(screen.queryByTestId("control-overlay")).not.toBeInTheDocument();
    },
  );

  it("keeps the loading state before the Worker is ready", () => {
    render(
      <AgentPanel
        paneId="pane-1"
        podKey="pod-1"
        isActive
        showHeader={false}
      />,
    );

    expect(screen.queryByTestId("agent-workspace")).not.toBeInTheDocument();
    expect(screen.getByText("Waiting for Pod to be ready...")).toBeInTheDocument();
  });
});
