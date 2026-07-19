import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TooltipProvider } from "@/components/ui/tooltip";

vi.mock("react-router-dom", async (load) => {
  const actual = await load<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigate };
});
vi.mock("@/hooks/useAvailableAgents", () => ({ useAvailableAgents: vi.fn() }));
vi.mock("@/hooks/useAgents", () => ({ useSessionAgent: vi.fn() }));
vi.mock("@/lib/workerSessionMutations", () => ({
  forkSnapshotSession: vi.fn(),
  forkWorkerSession: vi.fn(),
}));

import { useAvailableAgents } from "@/hooks/useAvailableAgents";
import { useSessionAgent } from "@/hooks/useAgents";
import { forkSnapshotSession, forkWorkerSession } from "@/lib/workerSessionMutations";
import { ForkSessionDialog } from "./ForkSessionDialog";

const navigate = vi.fn();
const source = { id: "codex-cli", name: "codex", harness: "codex" };
const agents = [
  {
    id: "codex-cli",
    workerTypeSlug: "codex-cli",
    supportedModes: ["acp", "pty"] as const,
    requiresModelResource: true,
    name: "codex",
    display_name: "Codex",
    description: null,
    harness: "codex",
    skills: [],
  },
  {
    id: "claude-code",
    workerTypeSlug: "claude-code",
    supportedModes: ["acp"] as const,
    requiresModelResource: false,
    name: "claude",
    display_name: "Claude",
    description: null,
    harness: "claude-sdk",
    skills: [],
  },
  {
    id: "custom_1",
    name: "Custom",
    display_name: "Custom",
    description: null,
    harness: "openai-agents",
    skills: [],
  },
];

function renderDialog(props: { upToResponseId?: string | null } = {}) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <TooltipProvider>
        <MemoryRouter>
          <ForkSessionDialog
            sourceSessionId="session_1"
            sourceTitle="Original"
            upToResponseId={props.upToResponseId}
            open
            onOpenChange={vi.fn()}
          />
        </MemoryRouter>
      </TooltipProvider>
    </QueryClientProvider>,
  );
}

function chooseAgent(id: string): void {
  const trigger = screen.getByTestId("fork-session-agent-select");
  fireEvent.pointerDown(trigger, new MouseEvent("pointerdown", { bubbles: true, button: 0 }));
  fireEvent.click(trigger);
  fireEvent.click(screen.getByTestId(`fork-session-agent-option-${id}`));
}

beforeEach(() => {
  navigate.mockReset();
  vi.mocked(useAvailableAgents).mockReturnValue({ data: agents } as ReturnType<typeof useAvailableAgents>);
  vi.mocked(useSessionAgent).mockReturnValue({ data: source } as ReturnType<typeof useSessionAgent>);
  vi.mocked(forkSnapshotSession).mockReset();
  vi.mocked(forkWorkerSession).mockReset();
  vi.mocked(forkSnapshotSession).mockResolvedValue({ id: "fork_1" });
  vi.mocked(forkWorkerSession).mockResolvedValue({ id: "fork_2" });
});

afterEach(cleanup);

describe("ForkSessionDialog", () => {
  it("keeps same-Agent forks as snapshot operations", async () => {
    renderDialog();
    fireEvent.change(screen.getByTestId("fork-session-title-input"), { target: { value: "Copy" } });
    fireEvent.click(screen.getByTestId("fork-session-submit"));

    await waitFor(() =>
      expect(forkSnapshotSession).toHaveBeenCalledWith({
        sourceId: "session_1",
        title: "Copy",
        upToResponseId: undefined,
      }),
    );
    expect(forkWorkerSession).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith("/c/fork_1");
  });

  it("uses the authoritative plan mutation for a cross-Agent fork", async () => {
    renderDialog({ upToResponseId: "response_1" });
    chooseAgent("claude-code");
    fireEvent.click(screen.getByTestId("fork-session-submit"));

    await waitFor(() =>
      expect(forkWorkerSession).toHaveBeenCalledWith({
        sourceId: "session_1",
        sourceAgentId: "codex-cli",
        agentId: "claude-code",
        workerTypeSlug: "claude-code",
        supportedModes: ["acp"],
        requiresModelResource: false,
        title: undefined,
        upToResponseId: "response_1",
      }),
    );
    expect(forkSnapshotSession).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith("/c/fork_2");
  });

  it("does not expose agents lacking authoritative Worker metadata", () => {
    renderDialog();
    fireEvent.pointerDown(screen.getByTestId("fork-session-agent-select"));

    expect(screen.queryByTestId("fork-session-agent-option-custom_1")).not.toBeInTheDocument();
  });

  it("shows a request failure without navigation", async () => {
    vi.mocked(forkSnapshotSession).mockRejectedValueOnce(new Error("plan changed"));
    renderDialog();
    fireEvent.click(screen.getByTestId("fork-session-submit"));

    expect(await screen.findByTestId("fork-session-error")).toHaveTextContent("plan changed");
    expect(navigate).not.toHaveBeenCalled();
  });
});
