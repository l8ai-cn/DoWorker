import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { ArtifactCard } from "../../ArtifactCard";
import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "../../contracts";
import { createBuiltinContentRenderers } from "../builtinContentRenderers";

describe("image edit artifact registration", () => {
  it("reports unsupported instead of guessing when the image_edit manifest is absent", () => {
    const runtime = createRuntime();
    render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={imageEditRoleWithoutManifest()}
        runtime={runtime}
        sessionId="session-1"
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "image_edit_manifest_missing",
    );
    expect(runtime.loadArtifact).not.toHaveBeenCalled();
  });

  it("shows artifact action failures instead of swallowing them", async () => {
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:source"),
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
    });
    const runtime = createRuntime();
    vi.mocked(runtime.loadArtifact!).mockResolvedValue(
      new Blob(["source"], { type: "image/png" }),
    );
    runtime.executeArtifactAction = vi.fn(async () => {
      throw new Error("图片编辑提交失败");
    });

    render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={editableImage()}
        runtime={runtime}
        sessionId="session-1"
      />,
    );

    fireEvent.change(
      await screen.findByRole("textbox", { name: "编辑说明" }),
      { target: { value: "移除背景" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "提交编辑" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "图片编辑提交失败",
    );
  });
});

function imageEditRoleWithoutManifest(): AgentArtifactItem {
  return {
    actions: ["image.edit"],
    artifactId: "image-1",
    filename: "result.png",
    grants: [],
    id: "artifact-image-1",
    kind: "artifact",
    manifest: null,
    mimeType: "image/png",
    representations: [],
    revision: 1n,
    role: "image_edit",
    schemaVersion: "1",
    selectedRepresentationId: null,
    status: "completed",
  };
}

function createRuntime(): AgentSessionRuntime {
  return {
    close: vi.fn(),
    getSnapshot: vi.fn(),
    interrupt: vi.fn(),
    loadArtifact: vi.fn(),
    loadOlder: vi.fn(),
    open: vi.fn(),
    resolvePermission: vi.fn(),
    sendMessage: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    updateConfiguration: vi.fn(),
  };
}

function editableImage(): AgentArtifactItem {
  return {
    ...imageEditRoleWithoutManifest(),
    grants: [
      {
        actions: ["image.edit"],
        grantId: "image-edit",
        representationIds: ["source"],
      },
    ],
    manifest: {
      annotations: [],
      candidateRepresentationIds: [],
      kind: "image_edit",
      regions: [],
      sourceDimensions: { height: 600, width: 800 },
      sourceRepresentationId: "source",
    },
    representations: [
      {
        mediaType: "image/png",
        representationId: "source",
        revision: 1n,
        status: "ready",
      },
    ],
    selectedRepresentationId: "source",
  };
}
