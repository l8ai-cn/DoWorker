import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import type { AgentToolActivityItem } from "../agentToolContracts";
import type { AgentSessionRuntime } from "../contracts";
import { ContentRendererRegistry } from "../registry/ContentRendererRegistry";
import { ToolRendererRegistry } from "../registry/ToolRendererRegistry";
import type { AgentContentRendererRegistration } from "./contentRendererTypes";
import type { AgentToolRendererRegistration } from "./rendererTypes";
import { ResultWorkbench } from "./ResultWorkbench";

const artifact: AgentArtifactItem = {
  actions: [],
  artifactId: "image-1",
  filename: "result.png",
  grants: [{
    actions: ["artifact.download"],
    grantId: "grant-download",
    representationIds: [],
  }],
  id: "artifact-1",
  kind: "artifact",
  manifest: null,
  mimeType: "image/png",
  representations: [],
  revision: 1n,
  role: "preview",
  schemaVersion: "1",
  selectedRepresentationId: null,
  status: "completed",
};
const tool: AgentToolActivityItem = {
  id: "tool-1",
  identity: {
    namespace: "agentsmesh.image",
    schemaVersion: "1",
    semanticKey: "edit",
  },
  inputValue: { instruction: "remove background" },
  kind: "tool",
  results: [],
  status: "completed",
  title: "Edit image",
};

beforeEach(() => {
  Object.defineProperty(URL, "createObjectURL", {
    configurable: true,
    value: vi.fn(() => "blob:result"),
  });
  Object.defineProperty(URL, "revokeObjectURL", {
    configurable: true,
    value: vi.fn(),
  });
});

describe("ResultWorkbench", () => {
  it("shows conversation and results together in wide containers", async () => {
    render(
      <ResultWorkbench
        artifacts={[artifact]}
        conversation={<div>Conversation surface</div>}
        mode="wide"
        runtime={runtime()}
        sessionId="session-1"
      />,
    );

    expect(screen.getByText("Conversation surface")).toBeVisible();
    expect(await screen.findByRole("img", { name: "result.png" })).toBeVisible();
  });

  it("defers result rendering until the narrow results tab is opened", async () => {
    const agentRuntime = runtime();
    render(
      <ResultWorkbench
        artifacts={[artifact]}
        conversation={<div>Conversation surface</div>}
        mode="narrow"
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(screen.getByText("Conversation surface")).toBeVisible();
    expect(agentRuntime.loadArtifact).not.toHaveBeenCalled();
    const resultsTab = screen.getByRole("tab", { name: "Results" });
    expect(
      document.getElementById(resultsTab.getAttribute("aria-controls")!),
    ).toBeEmptyDOMElement();

    fireEvent.click(resultsTab);
    const image = await screen.findByRole("img", { name: "result.png" });
    expect(screen.getByRole("button", { name: "result.png" })).toHaveTextContent(
      "result.png",
    );
    expect(image.closest("section")).toHaveAttribute("aria-hidden", "false");
    expect(image.closest("section")).not.toHaveClass("hidden");

    fireEvent.click(screen.getByRole("tab", { name: "Conversation" }));
    expect(
      screen.getByRole("img", { name: "result.png", hidden: true }),
    ).toBe(image);
    expect(image.closest("section")).toHaveAttribute("aria-hidden", "true");
  });

  it("supports roving focus with arrow keys in narrow tabs", () => {
    render(
      <ResultWorkbench
        artifacts={[artifact]}
        conversation={<div>Conversation surface</div>}
        mode="narrow"
        runtime={runtime()}
        sessionId="session-1"
      />,
    );
    const conversation = screen.getByRole("tab", { name: "Conversation" });
    const results = screen.getByRole("tab", { name: "Results" });
    conversation.focus();

    fireEvent.keyDown(conversation.closest('[role="tablist"]')!, {
      key: "ArrowRight",
    });

    expect(results).toHaveFocus();
    expect(conversation).toHaveAttribute("aria-controls");
    expect(results).toHaveAttribute("aria-controls");
  });

  it("triggers an exact tool workbench renderer in the results pane", () => {
    const toolRenderers =
      new ToolRendererRegistry<AgentToolRendererRegistration>();
    toolRenderers.register(
      tool.identity,
      { workbench: ({ item }) => <div>Workbench for {item.title}</div> },
      "test.image",
    );

    render(
      <ResultWorkbench
        artifacts={[]}
        conversation={<div>Conversation surface</div>}
        mode="wide"
        runtime={runtime()}
        sessionId="session-1"
        toolRenderers={toolRenderers}
        tools={[tool]}
      />,
    );

    expect(screen.getByText("Conversation surface")).toBeVisible();
    expect(screen.getByText("Workbench for Edit image")).toBeVisible();
  });

  it("does not guess a workbench renderer for another schema", () => {
    const toolRenderers =
      new ToolRendererRegistry<AgentToolRendererRegistration>();
    toolRenderers.register(
      { ...tool.identity, schemaVersion: "2" },
      { workbench: () => <div>Wrong workbench</div> },
      "test.image",
    );

    render(
      <ResultWorkbench
        artifacts={[]}
        conversation={<div>Conversation surface</div>}
        mode="narrow"
        runtime={runtime()}
        sessionId="session-1"
        toolRenderers={toolRenderers}
        tools={[tool]}
      />,
    );

    expect(screen.queryByText("Wrong workbench")).not.toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Results" })).not.toBeInTheDocument();
  });

  it("passes the runtime, session, and rich artifact metadata to content renderers", () => {
    const contentRenderers =
      new ContentRendererRegistry<AgentContentRendererRegistration>();
    contentRenderers.register(
      {
        blockKind: "artifact",
        mediaType: "image/png",
        role: "preview",
        schemaVersion: "1",
      },
      {
        viewer: ({ item, runtime: receivedRuntime, sessionId }) => (
          <div>
            {sessionId}:{item.revision.toString()}:
            {String(receivedRuntime === agentRuntime)}
          </div>
        ),
      },
      "test.image",
    );
    const agentRuntime = runtime();

    render(
      <ResultWorkbench
        artifacts={[artifact]}
        contentRenderers={contentRenderers}
        conversation={<div>Conversation surface</div>}
        mode="wide"
        runtime={agentRuntime}
        sessionId="session-rich"
      />,
    );

    expect(screen.getByText("session-rich:1:true")).toBeVisible();
  });
});

function runtime(): AgentSessionRuntime {
  return {
    close: vi.fn(),
    getSnapshot: vi.fn(),
    interrupt: vi.fn(),
    loadArtifact: vi.fn(async () => new Blob(["image"], { type: "image/png" })),
    loadOlder: vi.fn(),
    open: vi.fn(),
    resolvePermission: vi.fn(),
    sendMessage: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    updateConfiguration: vi.fn(),
  };
}
