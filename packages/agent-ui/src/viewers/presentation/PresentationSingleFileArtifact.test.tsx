import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { ArtifactCard } from "../../ArtifactCard";
import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "../../contracts";
import { createBuiltinContentRenderers } from "../builtinContentRenderers";

describe("single-file presentation artifacts", () => {
  it("keeps a PPTX on the generic file card when no image slide representation exists", async () => {
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:deck"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
    render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={singleFilePresentation()}
        runtime={runtime()}
        sessionId="session-1"
      />,
    );

    expect(await screen.findByText("PowerPoint")).toBeVisible();
    expect(screen.queryByText("演示文稿")).not.toBeInTheDocument();
  });
});

function singleFilePresentation(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "deck-1",
    filename: "deck.pptx",
    grants: [{
      actions: ["artifact.download"],
      grantId: "grant-download",
      representationIds: [],
    }],
    id: "artifact-deck-1",
    kind: "artifact",
    manifest: {
      deckRevision: 8n,
      kind: "presentation",
      selectedVersionId: "v1",
      slides: [
        {
          pageRepresentationId: "deck-source",
          position: 1,
          slideId: "slide-1",
          title: "项目概览",
        },
      ],
      versions: [{ id: "v1", label: "初稿", revision: 8n }],
    },
    mimeType:
      "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    representations: [
      {
        filename: "deck.pptx",
        mediaType:
          "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        representationId: "deck-source",
        revision: 8n,
        status: "ready",
      },
    ],
    revision: 8n,
    role: "presentation",
    schemaVersion: "1",
    selectedRepresentationId: "deck-source",
    status: "completed",
  };
}

function runtime(): AgentSessionRuntime {
  return {
    close: vi.fn(),
    getSnapshot: vi.fn(),
    interrupt: vi.fn(),
    loadArtifact: vi.fn(async () => new Blob(["deck"])),
    loadOlder: vi.fn(),
    open: vi.fn(),
    resolvePermission: vi.fn(),
    sendMessage: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    updateConfiguration: vi.fn(),
  };
}
