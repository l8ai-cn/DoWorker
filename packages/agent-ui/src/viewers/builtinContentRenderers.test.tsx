import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import type { AgentContentRendererProps } from "../react/contentRendererTypes";
import { createBuiltinContentRenderers } from "./builtinContentRenderers";

describe("createBuiltinContentRenderers", () => {
  it("registers the video viewer by exact content identity", async () => {
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:video"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
    const registry = createBuiltinContentRenderers();
    const registration = registry.lookup({
      blockKind: "artifact",
      mediaType: "video/mp4",
      role: "preview",
      schemaVersion: "1",
    });

    expect(registration?.viewer).toBeDefined();
    expect(
      registry.lookup({
        blockKind: "artifact",
        mediaType: "video/mp4",
        role: "preview",
        schemaVersion: "2",
      }),
    ).toBeUndefined();

    const Viewer = registration!.viewer;
    render(<Viewer {...videoProps()} />);
    expect(
      await screen.findByLabelText("视频预览：demo.mp4"),
    ).toBeVisible();
  });
});

function videoProps(): AgentContentRendererProps {
  const runtime = {
    close: () => undefined,
    getSnapshot: () => {
      throw new Error("unused");
    },
    interrupt: async () => undefined,
    loadArtifact: async () => new Blob(["video"], { type: "video/mp4" }),
    loadOlder: async () => undefined,
    open: async () => undefined,
    resolvePermission: async () => undefined,
    sendMessage: async () => undefined,
    subscribe: () => () => undefined,
    updateConfiguration: async () => undefined,
  };
  return {
    filename: "demo.mp4",
    item: {
      actions: [],
      artifactId: "video-1",
      filename: "demo.mp4",
      grants: [],
      id: "artifact-video-1",
      kind: "artifact",
      manifest: null,
      mimeType: "video/mp4",
      representations: [],
      revision: 1n,
      role: "preview",
      schemaVersion: "1",
      selectedRepresentationId: null,
      status: "completed",
    },
    runtime,
    sessionId: "session-1",
  };
}
