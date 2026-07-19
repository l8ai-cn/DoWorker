import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { ImportCodexDialog } from "./ImportCodexDialog";
import { useAvailableAgents } from "@/hooks/useAvailableAgents";
import { importWorkerSession } from "@/lib/workerSessionMutations";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigateMock };
});
vi.mock("@/hooks/useAvailableAgents", () => ({ useAvailableAgents: vi.fn() }));
vi.mock("@/lib/workerSessionMutations", () => ({ importWorkerSession: vi.fn() }));

const useAvailableAgentsMock = vi.mocked(useAvailableAgents);
const importWorkerSessionMock = vi.mocked(importWorkerSession);

beforeEach(() => {
  vi.clearAllMocks();
  useAvailableAgentsMock.mockReturnValue({
    data: [{
      id: "codex-cli",
      name: "codex-cli",
      display_name: "Codex",
      description: null,
      harness: "codex",
      skills: [],
      workerTypeSlug: "codex-cli",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
    }],
  } as ReturnType<typeof useAvailableAgents>);
});

afterEach(cleanup);

describe("ImportCodexDialog", () => {
  it("does not offer a custom session agent with no authoritative Worker metadata", () => {
    useAvailableAgentsMock.mockReturnValue({
      data: [{
        id: "agent_session_opaque",
        name: "custom-session-agent",
        display_name: "Custom session agent",
        description: null,
        harness: "codex",
        skills: [],
      }],
    } as ReturnType<typeof useAvailableAgents>);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter>
          <ImportCodexDialog open onOpenChange={vi.fn()} />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(screen.queryByTestId("agent-card-agent_session_opaque")).not.toBeInTheDocument();
    expect(screen.getByTestId("import-codex-unavailable")).toHaveTextContent(
      "Worker creation metadata is unavailable",
    );
  });

  it("imports through the authoritative Worker session mutation", async () => {
    importWorkerSessionMock.mockResolvedValue({ id: "conv_imported" });
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter>
          <ImportCodexDialog open onOpenChange={vi.fn()} />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    fireEvent.change(screen.getByTestId("import-codex-source-input"), {
      target: { value: "/tmp/rollout.jsonl" },
    });
    fireEvent.change(screen.getByTestId("import-codex-title-input"), {
      target: { value: "Imported transcript" },
    });
    fireEvent.click(screen.getByTestId("agent-card-codex-cli"));
    fireEvent.click(screen.getByTestId("import-codex-submit"));

    await waitFor(() => {
      expect(importWorkerSessionMock).toHaveBeenCalledWith({
        agentId: "codex-cli",
        sourcePath: "/tmp/rollout.jsonl",
        title: "Imported transcript",
        workerTypeSlug: "codex-cli",
        supportedModes: ["acp", "pty"],
        requiresModelResource: true,
      });
    });
    expect(navigateMock).toHaveBeenCalledWith("/c/conv_imported");
  });
});
