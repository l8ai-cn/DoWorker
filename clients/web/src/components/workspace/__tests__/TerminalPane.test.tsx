import { waitFor } from "@testing-library/react";
import { render, screen, fireEvent } from "@/test/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  addPane: vi.fn(),
  removePaneByPodKey: vi.fn(),
  wakePod: vi.fn(),
}));

vi.mock("@/hooks", () => ({
  usePodStatus: () => ({
    podStatus: "terminated",
    isPodReady: false,
    podError: "Pod terminated",
  }),
  useTerminal: () => ({
    terminalRef: { current: null },
    xtermRef: { current: null },
    connectionStatus: "connected",
    isRunnerDisconnected: false,
    syncSize: vi.fn(),
  }),
  useTouchScroll: vi.fn(),
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: (selector: (state: Record<string, unknown>) => unknown) => selector({
    initProgress: {},
    wakePod: mocks.wakePod,
  }),
}));

vi.mock("@/stores/workspace", () => ({
  useWorkspaceStore: (selector: (state: Record<string, unknown>) => unknown) => selector({
    terminalFontSize: 14,
    setActivePane: vi.fn(),
    splitPane: vi.fn(),
    panes: [],
    addPane: mocks.addPane,
    removePaneByPodKey: mocks.removePaneByPodKey,
  }),
}));

vi.mock("@/stores/autopilot", () => ({
  useAutopilotControllerByPodKey: () => null,
}));

vi.mock("../TerminalPaneHeader", () => ({
  TerminalPaneHeader: () => null,
}));

vi.mock("../RelayStatusOverlay", () => ({
  RelayStatusOverlay: () => null,
}));

vi.mock("../AutopilotOverlay", () => ({
  AutopilotOverlay: () => null,
}));

vi.mock("../AutopilotStartButton", () => ({
  AutopilotStartButton: () => null,
}));

vi.mock("../PodSelectorModal", () => ({
  PodSelectorModal: () => null,
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn() },
}));

import { TerminalPane } from "../TerminalPane";

describe("TerminalPane", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.wakePod.mockResolvedValue({ pod_key: "pod-resumed-789" });
  });

  it("replaces a terminated Worker pane with its resumed Worker", async () => {
    render(<TerminalPane paneId="pane-1" podKey="pod-terminated-123" isActive />);

    fireEvent.click(screen.getByRole("button", { name: "Wake Worker" }));

    await waitFor(() => {
      expect(mocks.wakePod).toHaveBeenCalledWith("pod-terminated-123");
    });
    expect(mocks.removePaneByPodKey).toHaveBeenCalledWith("pod-terminated-123");
    expect(mocks.addPane).toHaveBeenCalledWith("pod-resumed-789");
  });
});
