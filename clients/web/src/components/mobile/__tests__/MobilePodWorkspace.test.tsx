import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MobilePodWorkspace } from "../MobilePodWorkspace";

const mockFetchPod = vi.fn();
let mockPod: Record<string, unknown> | undefined;

vi.mock("@/stores/pod", () => ({
  usePod: () => mockPod,
  usePodStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ fetchPod: mockFetchPod }),
}));

vi.mock("@/components/workspace/TerminalPane", () => ({
  TerminalPane: ({
    podKey,
    showHeader,
    controlClientLabel,
  }: {
    podKey: string;
    showHeader: boolean;
    controlClientLabel: string;
  }) => (
    <div
      data-testid="terminal-pane"
      data-pod-key={podKey}
      data-show-header={String(showHeader)}
      data-control-client={controlClientLabel}
    />
  ),
}));

vi.mock("@/components/workspace/AgentPanel", () => ({
  AgentPanel: ({
    podKey,
    showHeader,
    controlClientLabel,
  }: {
    podKey: string;
    showHeader: boolean;
    controlClientLabel: string;
  }) => (
    <div
      data-testid="agent-panel"
      data-pod-key={podKey}
      data-show-header={String(showHeader)}
      data-control-client={controlClientLabel}
    />
  ),
}));

describe("MobilePodWorkspace", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFetchPod.mockResolvedValue(undefined);
    mockPod = undefined;
  });

  it("uses the terminal pane for non-ACP pods", () => {
    mockPod = { pod_key: "pod-pty", interaction_mode: "pty" };

    render(<MobilePodWorkspace podKey="pod-pty" />);

    expect(screen.getByTestId("terminal-pane")).toHaveAttribute("data-pod-key", "pod-pty");
    expect(screen.getByTestId("terminal-pane")).toHaveAttribute("data-show-header", "false");
    expect(screen.getByTestId("terminal-pane")).toHaveAttribute("data-control-client", "mobile");
    expect(screen.queryByTestId("agent-panel")).not.toBeInTheDocument();
  });

  it("uses the ACP panel for ACP pods", () => {
    mockPod = { pod_key: "pod-acp", interaction_mode: "acp" };

    render(<MobilePodWorkspace podKey="pod-acp" />);

    expect(screen.getByTestId("agent-panel")).toHaveAttribute("data-pod-key", "pod-acp");
    expect(screen.getByTestId("agent-panel")).toHaveAttribute("data-show-header", "false");
    expect(screen.getByTestId("agent-panel")).toHaveAttribute("data-control-client", "mobile");
    expect(screen.queryByTestId("terminal-pane")).not.toBeInTheDocument();
  });

  it("fetches the pod when the mobile route is opened from a cold cache", async () => {
    render(<MobilePodWorkspace podKey="pod-cold" />);

    await waitFor(() => expect(mockFetchPod).toHaveBeenCalledWith("pod-cold"));
    expect(screen.getByTestId("mobile-pod-loading")).toBeInTheDocument();
  });
});
