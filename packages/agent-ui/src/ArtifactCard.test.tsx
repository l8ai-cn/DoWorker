import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import { ActivityTimeline } from "./ActivityTimeline";
import { AgentWorkspace } from "./AgentWorkspace";
import type { AgentContentRendererRegistration } from "./react/contentRendererTypes";
import { ContentRendererRegistry } from "./registry/ContentRendererRegistry";
import {
  STATIC_HTML_CSP,
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
} from "./security/staticHtmlProfile";
import type {
  AgentArtifactItem,
  AgentSessionRuntime,
  AgentSessionSnapshot,
} from "./contracts";

const createObjectURL = vi.fn(() => "blob:artifact-preview");
const revokeObjectURL = vi.fn();

function runtime(loadArtifact: (sessionId: string, artifactId: string) => Promise<Blob>) {
  return {
    open: vi.fn(async () => undefined),
    close: vi.fn(),
    getSnapshot: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    sendMessage: vi.fn(async () => undefined),
    interrupt: vi.fn(async () => undefined),
    resolvePermission: vi.fn(async () => undefined),
    updateConfiguration: vi.fn(async () => undefined),
    loadOlder: vi.fn(async () => undefined),
    loadArtifact: vi.fn(loadArtifact),
  } as AgentSessionRuntime;
}

function artifact(
  patch: Partial<AgentArtifactItem> = {},
): AgentArtifactItem {
  return {
    actions: [],
    id: "artifact-item-1",
    kind: "artifact",
    artifactId: "artifact-1",
    filename: "report.pdf",
    grants: [],
    manifest: null,
    mimeType: "application/pdf",
    representations: [],
    revision: 1n,
    role: "preview",
    schemaVersion: "1",
    selectedRepresentationId: null,
    status: "completed",
    ...patch,
  };
}

describe("ArtifactCard", () => {
  beforeEach(() => {
    createObjectURL.mockClear();
    revokeObjectURL.mockClear();
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: createObjectURL,
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: revokeObjectURL,
    });
  });

  it("uses a content viewer only for the exact content identity", async () => {
    const renderers =
      new ContentRendererRegistry<AgentContentRendererRegistration>();
    renderers.register(
      {
        blockKind: "artifact",
        mediaType: "video/mp4",
        role: "preview",
        schemaVersion: "1",
      },
      {
        viewer: ({ filename }) => <div>专用视频查看器：{filename}</div>,
      },
      "test.video",
    );
    const agentRuntime = runtime(async () =>
      new Blob(["video"], { type: "video/mp4" }),
    );

    const { rerender } = render(
      <ActivityTimeline
        contentRenderers={renderers}
        items={[
          artifact({
            artifactId: "video-1",
            filename: "demo.mp4",
            mimeType: "video/mp4",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(await screen.findByText("专用视频查看器：demo.mp4")).toBeVisible();

    rerender(
      <ActivityTimeline
        contentRenderers={renderers}
        items={[
          artifact({
            artifactId: "video-2",
            filename: "demo-v2.mp4",
            mimeType: "video/mp4",
            schemaVersion: "2",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(
      screen.queryByText("专用视频查看器：demo-v2.mp4"),
    ).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Load demo-v2.mp4" }),
    );
    expect(await screen.findByLabelText("Video preview for demo-v2.mp4")).toBeVisible();
  });

  it("loads image blobs through the runtime and cleans the preview URL", async () => {
    const blob = new Blob(["image"], { type: "image/png" });
    const agentRuntime = runtime(async () => blob);
    const { unmount } = render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "image-1",
            filename: "diagram.png",
            mimeType: "image/png",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    const image = await screen.findByRole("img", { name: "diagram.png" });
    expect(image).toHaveAttribute("src", "blob:artifact-preview");
    expect(agentRuntime.loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "image-1",
    );
    expect(createObjectURL).toHaveBeenCalledWith(blob);

    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:artifact-preview");
  });

  it("receives the artifact runtime and session id through AgentWorkspace", async () => {
    const agentRuntime = runtime(async () =>
      new Blob(["image"], { type: "image/png" }),
    );
    const snapshot: AgentSessionSnapshot = {
      sessionId: "workspace-session",
      title: "Artifact review",
      agentLabel: "Codex",
      status: "completed",
      connection: "connected",
      interactionMode: "acp",
      capabilities: {
        sendMessage: true,
        interrupt: true,
        resolvePermission: true,
        updateConfiguration: true,
        terminal: false,
      },
      items: [
        artifact({
          artifactId: "workspace-image",
          filename: "workspace.png",
          mimeType: "image/png",
        }),
      ],
      plan: [],
      permissions: [],
      terminals: [],
      hasOlderItems: false,
      error: null,
    };
    vi.mocked(agentRuntime.getSnapshot).mockReturnValue(snapshot);

    render(
      <AgentWorkspace
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    fireEvent.click(await screen.findByRole("tab", { name: "Results" }));
    expect(
      await screen.findByRole("img", { name: "workspace.png" }),
    ).toBeVisible();
    expect(agentRuntime.loadArtifact).toHaveBeenCalledWith(
      "workspace-session",
      "workspace-image",
    );
  });

  it("renders video blobs as an inline preview", async () => {
    const agentRuntime = runtime(async () =>
      new Blob(["video"], { type: "video/mp4" }),
    );

    render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "video-1",
            filename: "demo.mp4",
            mimeType: "video/mp4",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Load demo.mp4" }));
    expect(
      await screen.findByLabelText("Video preview for demo.mp4"),
    ).toHaveAttribute("src", "blob:artifact-preview");
  });

  it("renders generated source files as an inline code preview", async () => {
    const source = "const title = 'Agent Workspace';\nconsole.log(title);";
    const agentRuntime = runtime(async () =>
      new Blob([source], { type: "text/javascript" }),
    );

    render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "source-1",
            filename: "generate-assets.mjs",
            mimeType: "text/javascript",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(
      await screen.findByLabelText("Preview generate-assets.mjs"),
    ).toHaveTextContent("const title = 'Agent Workspace'");
    expect(screen.getByText("Code file")).toBeVisible();
  });

  it("renders generated HTML inside the static sandbox profile", async () => {
    const agentRuntime = runtime(async () =>
      new Blob(["<button>重新开始</button><script>window.ready=true</script>"], {
        type: "text/html",
      }),
    );

    render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "html-1",
            filename: "index.html",
            mimeType: "text/html",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    const preview = await screen.findByTitle("Preview index.html");
    expect(preview).toHaveAttribute(
      "srcdoc",
      expect.stringContaining("<button>重新开始</button>"),
    );
    expect(preview).not.toHaveAttribute("src");
    expect(preview).toHaveAttribute("sandbox", STATIC_HTML_SANDBOX);
    expect(preview).toHaveAttribute(
      "referrerpolicy",
      STATIC_HTML_REFERRER_POLICY,
    );
    expect(preview.getAttribute("srcdoc")).toContain(STATIC_HTML_CSP);
    expect(
      screen.queryByRole("link", { name: "Open index.html" }),
    ).not.toBeInTheDocument();
  });

  it("shows file type with accessible open and download actions", async () => {
    const agentRuntime = runtime(async () =>
      new Blob(["slides"], {
        type: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
      }),
    );

    render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "slides-1",
            filename: "release-review.pptx",
            mimeType:
              "application/vnd.openxmlformats-officedocument.presentationml.presentation",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(await screen.findByText("PowerPoint")).toBeVisible();
    expect(
      screen.getByRole("link", { name: "Open release-review.pptx" }),
    ).toHaveAttribute("href", "blob:artifact-preview");
    expect(
      screen.getByRole("link", { name: "Download release-review.pptx" }),
    ).toHaveAttribute("download", "release-review.pptx");
  });

  it("keeps active document artifacts download-only", async () => {
    const agentRuntime = runtime(async () =>
      new Blob(["<svg><script>alert(1)</script></svg>"], {
        type: "image/svg+xml",
      }),
    );

    render(
      <ActivityTimeline
        items={[
          artifact({
            artifactId: "active-svg",
            filename: "diagram.svg",
            mimeType: "image/svg+xml",
          }),
        ]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(await screen.findByText("SVG document")).toBeVisible();
    expect(screen.queryByRole("img", { name: "diagram.svg" })).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Open diagram.svg" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Download diagram.svg" }),
    ).toHaveAttribute("download", "diagram.svg");
  });

  it("shows an explicit error when artifact loading fails", async () => {
    const agentRuntime = runtime(async () => {
      throw new Error("Artifact storage is unavailable");
    });

    render(
      <ActivityTimeline
        items={[artifact()]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(
      await screen.findByRole("alert"),
    ).toHaveTextContent("Artifact loading failed. Try again.");
    expect(createObjectURL).not.toHaveBeenCalled();
  });

  it("does not load failed artifact items", async () => {
    const agentRuntime = runtime(async () => new Blob(["unused"]));

    render(
      <ActivityTimeline
        items={[artifact({ status: "failed" })]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Artifact generation failed",
    );
    await waitFor(() =>
      expect(agentRuntime.loadArtifact).not.toHaveBeenCalled(),
    );
  });
});
