import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { vi } from "vitest";

import { AgentWorkspace } from "./AgentWorkspace";
import {
  agentWorkspaceRuntime as runtime,
  agentWorkspaceSnapshot as sessionSnapshot,
} from "./AgentWorkspace.test-fixture";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import { ToolRendererRegistry } from "./registry/ToolRendererRegistry";

describe("AgentWorkspace", () => {
  beforeEach(() => {
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:artifact"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
  });

  it("renders an intent-first empty workspace with real session capabilities", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByRole("heading", {
        name: "Codex, what should we work on?",
      }),
    ).toBeVisible();
    expect(screen.getByText("Agentic")).toBeVisible();
    expect(screen.getAllByText("Codex")).not.toHaveLength(0);
    expect(screen.getByText("Approvals")).toBeVisible();
    expect(screen.getAllByText("Terminal")).not.toHaveLength(0);
    expect(screen.getByLabelText("Message the agent")).toHaveAttribute(
      "placeholder",
      "Ask Codex to work on a task…",
    );
  });

  it("hides terminal controls when the runtime has no terminal capability", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.capabilities.terminal = false;
    snapshot.terminals = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    await screen.findByText("Release audit");
    expect(screen.queryByRole("tab", { name: "Terminal" })).not.toBeInTheDocument();
    expect(screen.queryByText("Terminal")).not.toBeInTheDocument();
  });

  it("renders the running agent, plan, activity, approval, and terminal resource", async () => {
    const snapshot = sessionSnapshot();
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(await screen.findByText("Release audit")).toBeVisible();
    expect(screen.getAllByText("Codex")).not.toHaveLength(0);
    expect(screen.getByText("Run verification")).toBeVisible();
    expect(screen.getAllByText("shell")).not.toHaveLength(0);
    expect(screen.getByTestId("unsupported-tool-preview")).toBeInTheDocument();
    expect(screen.getByText("Run release command")).toBeVisible();
    expect(screen.getByRole("tab", { name: "Terminal" })).toBeEnabled();
  });

  it("moves artifacts into a persistent results surface", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.items.push({
      actions: [],
      artifactId: "artifact-1",
      filename: "result.png",
      grants: [{
        actions: ["artifact.download"],
        grantId: "grant-download",
        representationIds: [],
      }],
      id: "artifact-item-1",
      kind: "artifact",
      manifest: null,
      mimeType: "image/png",
      representations: [],
      revision: 1n,
      role: "preview",
      schemaVersion: "1",
      selectedRepresentationId: null,
      status: "completed",
    });
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace runtime={agentRuntime} sessionId={snapshot.sessionId} />,
    );

    expect(await screen.findByRole("tab", { name: "Results" })).toBeVisible();
    expect(
      screen.queryByRole("img", { name: "result.png" }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "Results" }));

    expect(await screen.findByRole("img", { name: "result.png" })).toBeVisible();
    expect(
      screen.getByText("Used shell 1 time").closest("section"),
    ).toHaveAttribute("aria-hidden", "true");
  });

  it("sends messages and resolves approvals through the runtime", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.hasOlderItems = true;
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    fireEvent.change(await screen.findByLabelText("Message the agent"), {
      target: { value: "Check the changelog" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Send message" }));
    fireEvent.click(screen.getByRole("button", { name: "Approve" }));
    fireEvent.click(screen.getByRole("button", { name: "Load earlier activity" }));

    await waitFor(() => {
      expect(agentRuntime.sendMessage).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { text: "Check the changelog" },
      );
      expect(agentRuntime.resolvePermission).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        "permission-1",
        { action: "accept", content: { answers: {} } },
      );
      expect(agentRuntime.loadOlder).toHaveBeenCalledWith("session-1");
    });
  });

  it("renders assistant markdown and structured tool evidence", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [
      {
        id: "assistant-1",
        kind: "message",
        role: "assistant",
        text: "### Result\n\n```ts\nexport const ready = true;\n```",
        status: "completed",
      },
      {
        id: "tool-1",
        identity: {
          namespace: "agentsmesh.acp",
          schemaVersion: "1",
          semanticKey: "shell",
        },
        kind: "tool",
        results: [],
        title: "shell",
        input: "{}",
        output: "12 tests passed",
        status: "completed",
      },
    ];
    const { agentRuntime } = runtime(snapshot);

    render(<AgentWorkspace runtime={agentRuntime} sessionId={snapshot.sessionId} />);

    expect(await screen.findByRole("heading", { name: "Result" })).toBeVisible();
    expect(screen.getByText("export const ready = true;")).toBeVisible();
    expect(screen.queryByText("```ts")).not.toBeInTheDocument();
    const toolGroup = screen.getByText("Used shell 1 time").closest("details");
    expect(toolGroup).not.toHaveAttribute("open");
    for (const output of screen.getAllByText("12 tests passed")) {
      expect(output).not.toBeVisible();
    }
    fireEvent.click(screen.getByText("Used shell 1 time"));
    fireEvent.click(within(toolGroup!).getByText("Details"));
    expect(screen.getAllByText("12 tests passed")[0]).toBeVisible();
    expect(screen.queryByText("{}")).not.toBeInTheDocument();
  });

  it("localizes structured task activity without changing its evidence", async () => {
    const snapshot = sessionSnapshot();
    const shellIdentity = {
      namespace: "agentsmesh.acp",
      schemaVersion: "1",
      semanticKey: "shell",
    };
    snapshot.status = "idle";
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.items = [
      {
        id: "tool-zh",
        identity: shellIdentity,
        kind: "tool",
        results: [],
        title: "shell",
        input: JSON.stringify({ command: "pnpm test" }),
        output: "12 tests passed",
        status: "running",
      },
    ];
    const { agentRuntime } = runtime(snapshot);
    const toolRenderers =
      new ToolRendererRegistry<AgentToolRendererRegistration>();
    toolRenderers.register(
      shellIdentity,
      { presentation: { label: "Command" } },
      "test.shell",
    );

    render(
      <AgentWorkspace
        locale="zh-CN"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
        toolRenderers={toolRenderers}
      />,
    );

    const summary = await screen.findByText("运行了 1 个命令");
    const group = summary.closest("details");
    const groupSummary = group!.querySelector("summary")!;

    expect(group).not.toHaveAttribute("open");
    expect(within(groupSummary).getByText("执行中")).toBeVisible();
    expect(screen.getByText(/pnpm test/)).not.toBeVisible();

    fireEvent.click(summary);
    fireEvent.click(within(group!).getByText("详细信息"));

    expect(group).toHaveAttribute("open");
    expect(screen.getByText("详细信息")).toBeVisible();
    expect(screen.getByText(/pnpm test/)).toBeVisible();
  });

  it("shows command failures inside the workspace", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);
    vi.mocked(agentRuntime.sendMessage).mockRejectedValueOnce(
      new Error("Worker rejected the prompt"),
    );

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    fireEvent.change(await screen.findByLabelText("Message the agent"), {
      target: { value: "Run it" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Send message" }));

    expect(await screen.findByText("Worker rejected the prompt")).toBeVisible();
  });

  it("sends with Enter and preserves Shift+Enter", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    const input = await screen.findByLabelText("Message the agent");
    fireEvent.change(input, { target: { value: "First line" } });
    fireEvent.keyDown(input, { key: "Enter", shiftKey: true });
    expect(agentRuntime.sendMessage).not.toHaveBeenCalled();

    fireEvent.keyDown(input, { key: "Enter" });
    await waitFor(() =>
      expect(agentRuntime.sendMessage).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { text: "First line" },
      ),
    );
  });

  it("keeps running sessions interrupt-only even when a draft is present", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "running";
    snapshot.items = [];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    const input = await screen.findByLabelText("Message the agent");
    fireEvent.change(input, { target: { value: "Queue another prompt" } });

    expect(screen.getByRole("button", { name: "Stop agent" })).toBeVisible();
    fireEvent.keyDown(input, { key: "Enter" });
    expect(agentRuntime.sendMessage).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Stop agent" }));
    await waitFor(() =>
      expect(agentRuntime.interrupt).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
      ),
    );
  });

  it("keeps background tools interrupt-only after the session becomes idle", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "idle";
    snapshot.items = [
      {
        id: "tool-background",
        identity: {
          namespace: "agentsmesh.acp",
          schemaVersion: "1",
          semanticKey: "shell",
        },
        kind: "tool",
        results: [],
        title: "shell",
        input: "long-running command",
        status: "running",
      },
    ];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        terminalRuntime={terminalRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    const input = await screen.findByLabelText("Message the agent");
    fireEvent.change(input, { target: { value: "Do not queue this prompt" } });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(agentRuntime.sendMessage).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Stop agent" }));
    await waitFor(() =>
      expect(agentRuntime.interrupt).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
      ),
    );
  });
});
