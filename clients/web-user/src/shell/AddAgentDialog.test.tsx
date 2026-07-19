import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { AddAgentDialog } from "./AddAgentDialog";
import { useAvailableAgents, type AvailableAgent } from "@/hooks/useAvailableAgents";
import { createWorkerSession } from "@/lib/workerSessionMutations";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigateMock };
});
vi.mock("@/hooks/useAvailableAgents", () => ({ useAvailableAgents: vi.fn() }));
vi.mock("@/lib/workerSessionMutations", () => ({ createWorkerSession: vi.fn() }));

const useAvailableAgentsMock = vi.mocked(useAvailableAgents);
const createWorkerSessionMock = vi.mocked(createWorkerSession);

const AGENTS: AvailableAgent[] = [
  {
    id: "ag_claude",
    name: "claude-native-ui",
    display_name: "Claude Code",
    description: "Claude Code agent",
    harness: "claude-native",
    skills: [],
    workerTypeSlug: "claude-native-ui",
    supportedModes: ["acp", "pty"],
    requiresModelResource: true,
  },
  {
    id: "ag_codex",
    name: "codex",
    display_name: "codex",
    description: null,
    harness: "codex",
    skills: [],
    workerTypeSlug: "codex",
    supportedModes: ["acp", "pty"],
    requiresModelResource: true,
  },
];

function mockAgents(agents: AvailableAgent[]) {
  useAvailableAgentsMock.mockReturnValue({
    data: agents,
  } as unknown as ReturnType<typeof useAvailableAgents>);
}

function renderDialog(parentSessionId = "conv_parent") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const invalidateSpy = vi.spyOn(client, "invalidateQueries");
  const utils = render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AddAgentDialog parentSessionId={parentSessionId} open onOpenChange={vi.fn()} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
  return { ...utils, invalidateSpy };
}

beforeEach(() => {
  useAvailableAgentsMock.mockReset();
  createWorkerSessionMock.mockReset();
  navigateMock.mockReset();
  mockAgents(AGENTS);
});

afterEach(cleanup);

describe("AddAgentDialog", () => {
  it("lists the available agents from the catalog", () => {
    renderDialog();
    expect(screen.getByTestId("agent-card-ag_claude")).toHaveTextContent("Claude Code");
    expect(screen.getByTestId("agent-card-ag_codex")).toHaveTextContent("codex");
  });

  it("does not offer session-derived agents without Worker creation metadata", () => {
    mockAgents([
      ...AGENTS,
      {
        id: "agent_session_opaque",
        name: "custom-session-agent",
        display_name: "Custom session agent",
        description: null,
        harness: "codex",
        skills: [],
      },
    ]);

    renderDialog();

    expect(screen.queryByTestId("agent-card-agent_session_opaque")).not.toBeInTheDocument();
    expect(screen.getByTestId("add-agent-unavailable")).toHaveTextContent(
      "Worker creation metadata is unavailable",
    );
  });

  it("submits ui:<agent>:<name> with the parent link and a null sub_agent_name", async () => {
    createWorkerSessionMock.mockResolvedValue({
      id: "conv_child",
    });

    const { invalidateSpy } = renderDialog("conv_parent");

    fireEvent.click(screen.getByTestId("agent-card-ag_claude"));
    // Nothing is prefilled — the user types the name themselves.
    fireEvent.change(screen.getByTestId("add-agent-name-input"), {
      target: { value: "jimmy" },
    });
    fireEvent.click(screen.getByTestId("add-agent-submit"));

    await waitFor(() => expect(createWorkerSessionMock).toHaveBeenCalledTimes(1));
    // Whole call asserted: the 3-segment title carries the typed name, the
    // parent link, and sub_agent_name=null (so the runner resolves the
    // child's own agent_id).
    expect(createWorkerSessionMock).toHaveBeenCalledWith({
      agentId: "ag_claude",
      initialItems: [],
      parentSessionId: "conv_parent",
      subAgentName: null,
      title: "ui:claude-native-ui:jimmy",
      workerTypeSlug: "claude-native-ui",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
    });
    // Rail refreshed for the parent, then navigated into the new child.
    await waitFor(() => expect(navigateMock).toHaveBeenCalledWith("/c/conv_child"));
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ["conversation", "conv_parent", "child_sessions"],
    });
  });

  it("starts the name empty and blocks submit until the user types one", async () => {
    createWorkerSessionMock.mockResolvedValue({
      id: "conv_child",
    });
    renderDialog("conv_parent");

    fireEvent.click(screen.getByTestId("agent-card-ag_codex"));
    // Empty by default — the user must name the agent themselves.
    const input = screen.getByTestId("add-agent-name-input");
    expect(input).toHaveValue("");
    expect(screen.getByTestId("add-agent-submit")).toBeDisabled();

    // Once named, submit enables and the title carries the name verbatim.
    fireEvent.change(input, { target: { value: "reviewer" } });
    expect(screen.getByTestId("add-agent-submit")).toBeEnabled();
    fireEvent.click(screen.getByTestId("add-agent-submit"));

    await waitFor(() => expect(createWorkerSessionMock).toHaveBeenCalledTimes(1));
    expect(createWorkerSessionMock).toHaveBeenCalledWith({
      agentId: "ag_codex",
      initialItems: [],
      parentSessionId: "conv_parent",
      subAgentName: null,
      title: "ui:codex:reviewer",
      workerTypeSlug: "codex",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
    });
  });

  // A planned feature wants the user to task the newly-added Codex reviewer at
  // creation time (e.g. "review the implementation against the design").
  // The dialog has no initial-prompt field yet, so it always sends []
  // initial items and the child opens untasked. `it.fails` is the strict
  // tripwire: the body fails today (no such field to type into), and goes
  // red the moment a prompt field lands and its text flows into
  // createSession — at which point promote this to a normal assertion.
  it.fails("seeds the user's initial review prompt into the child transcript", async () => {
    createWorkerSessionMock.mockResolvedValue({
      id: "conv_child",
    });
    renderDialog("conv_parent");

    fireEvent.click(screen.getByTestId("agent-card-ag_codex"));
    fireEvent.change(screen.getByTestId("add-agent-name-input"), {
      target: { value: "reviewer" },
    });
    // No initial-prompt field exists today — getByTestId throws, which is
    // the expected failure that keeps this xfail-equivalent green.
    fireEvent.change(screen.getByTestId("add-agent-initial-prompt-input"), {
      target: { value: "review the implementation against designs/feature-x.md" },
    });
    fireEvent.click(screen.getByTestId("add-agent-submit"));

    await waitFor(() => expect(createWorkerSessionMock).toHaveBeenCalledTimes(1));
    // The prompt must travel as initial_items (a seeded user message), not
    // the empty [] the dialog sends today.
    const request = createWorkerSessionMock.mock.calls[0][0];
    expect(request.initialItems).not.toEqual([]);
    expect(JSON.stringify(request.initialItems)).toContain("designs/feature-x.md");
  });

  it("shows an empty-state and a disabled submit when no agents are available", () => {
    mockAgents([]);
    renderDialog();
    expect(screen.getByTestId("add-agent-empty")).toBeInTheDocument();
    expect(screen.getByTestId("add-agent-submit")).toBeDisabled();
  });

  it("surfaces the server error inline on failure and does not navigate", async () => {
    createWorkerSessionMock.mockRejectedValue(new Error("409 label already in use"));
    renderDialog();

    fireEvent.click(screen.getByTestId("agent-card-ag_codex"));
    fireEvent.change(screen.getByTestId("add-agent-name-input"), {
      target: { value: "reviewer" },
    });
    fireEvent.click(screen.getByTestId("add-agent-submit"));

    await waitFor(() =>
      expect(screen.getByTestId("add-agent-error")).toHaveTextContent("409 label already in use"),
    );
    // A failed create must not navigate the user away from the parent.
    expect(navigateMock).not.toHaveBeenCalled();
  });
});
