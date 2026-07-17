import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "../contracts";
import { ArtifactCard } from "../ArtifactCard";
import { createBuiltinContentRenderers } from "./builtinContentRenderers";

const createObjectURL = vi.fn(
  (blob: Blob) => `blob:${blob.type}:${createObjectURL.mock.calls.length}`,
);
const revokeObjectURL = vi.fn();

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

describe("rich built-in artifact renderers", () => {
  it("loads source and result representations, compares them, and submits an image edit action", async () => {
    const runtime = artifactRuntime();
    const randomUUID = vi
      .spyOn(globalThis.crypto, "randomUUID")
      .mockReturnValue("00000000-0000-4000-8000-000000000001");
    const { unmount } = render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={imageEditArtifact()}
        runtime={runtime}
        sessionId="session-1"
      />,
    );

    expect(
      await screen.findByRole("button", { name: "并排比较" }),
    ).toBeVisible();
    expect(runtime.loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "image-1",
      "source",
    );
    expect(runtime.loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "image-1",
      "result",
    );

    fireEvent.click(screen.getByRole("button", { name: "并排比较" }));
    expect(screen.getByRole("img", { name: "源图" })).toBeVisible();
    expect(screen.getByRole("img", { name: "结果图" })).toBeVisible();

    fireEvent.change(screen.getByRole("textbox", { name: "编辑说明" }), {
      target: { value: "删除选区中的文字" },
    });
    fireEvent.click(screen.getByRole("button", { name: "提交编辑" }));

    expect(runtime.executeArtifactAction).toHaveBeenCalledWith("session-1", {
      actionSchemaVersion: "1",
      actionType: "image.edit",
      artifactId: "image-1",
      baseRevision: 12n,
      commandId: "00000000-0000-4000-8000-000000000001",
      payload: {
        instruction: "删除选区中的文字",
        normalizedRegion: {
          height: 0.4,
          width: 0.3,
          x: 0.1,
          y: 0.2,
        },
        sourceDimensions: { height: 1080, width: 1920 },
      },
      representationId: "source",
    });

    unmount();
    expect(revokeObjectURL).toHaveBeenCalledTimes(2);
    randomUUID.mockRestore();
  });

  it("loads only the active video version and poster until another version is selected", async () => {
    const runtime = artifactRuntime();
    render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={videoArtifact()}
        runtime={runtime}
        sessionId="session-1"
      />,
    );

    expect(
      await screen.findByLabelText("视频预览：playable.mp4"),
    ).toBeVisible();
    expect(screen.getByText("时长 1:05")).toBeVisible();
    expect(screen.getByRole("combobox", { name: "选择视频版本" })).toHaveValue(
      "playable",
    );
    for (const representationId of ["playable", "poster"]) {
      expect(runtime.loadArtifact).toHaveBeenCalledWith(
        "session-1",
        "video-1",
        representationId,
      );
    }
    expect(runtime.loadArtifact).not.toHaveBeenCalledWith(
      "session-1",
      "video-1",
      "original",
    );
    expect(runtime.loadArtifact).not.toHaveBeenCalledWith(
      "session-1",
      "video-1",
      "derivative",
    );

    fireEvent.change(
      screen.getByRole("combobox", { name: "选择视频版本" }),
      { target: { value: "original" } },
    );
    await waitFor(() =>
      expect(runtime.loadArtifact).toHaveBeenCalledWith(
        "session-1",
        "video-1",
        "original",
      ),
    );
  });

  it("loads presentation pages and thumbnails and sends slide actions through the shared runtime", async () => {
    const runtime = artifactRuntime();
    const randomUUID = vi
      .spyOn(globalThis.crypto, "randomUUID")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000002")
      .mockReturnValueOnce("00000000-0000-4000-8000-000000000003");
    render(
      <ArtifactCard
        contentRenderers={createBuiltinContentRenderers()}
        item={presentationArtifact()}
        runtime={runtime}
        sessionId="session-1"
      />,
    );

    expect(
      await screen.findByRole("img", { name: "第 1 页：项目概览" }),
    ).toBeVisible();
    expect(runtime.loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "deck-1",
      "page-1",
    );
    expect(runtime.loadArtifact).toHaveBeenCalledWith(
      "session-1",
      "deck-1",
      "thumb-1",
    );

    fireEvent.click(screen.getByRole("button", { name: "重新生成当前页" }));
    expect(runtime.executeArtifactAction).toHaveBeenCalledWith("session-1", {
      actionSchemaVersion: "1",
      actionType: "presentation.regenerate_slide",
      artifactId: "deck-1",
      baseRevision: 8n,
      commandId: "00000000-0000-4000-8000-000000000002",
      payload: { slideId: "slide-1" },
      representationId: "page-1",
    });
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "重新生成当前页" }),
      ).toBeEnabled(),
    );
    fireEvent.change(
      screen.getByRole("combobox", { name: "选择演示文稿版本" }),
      { target: { value: "v2" } },
    );
    expect(runtime.executeArtifactAction).toHaveBeenNthCalledWith(
      2,
      "session-1",
      {
        actionSchemaVersion: "1",
        actionType: "presentation.select_version",
        artifactId: "deck-1",
        baseRevision: 8n,
        commandId: "00000000-0000-4000-8000-000000000003",
        payload: { versionId: "v2" },
        representationId: "deck-source",
      },
    );
    await waitFor(() =>
      expect(
        screen.getByRole("combobox", { name: "选择演示文稿版本" }),
      ).toBeEnabled(),
    );
    randomUUID.mockRestore();
  });
});

function artifactRuntime(): AgentSessionRuntime {
  return {
    close: vi.fn(),
    executeArtifactAction: vi.fn(async () => undefined),
    getSnapshot: vi.fn(),
    interrupt: vi.fn(),
    loadArtifact: vi.fn(async (_sessionId, _artifactId, representationId) => {
      const type = representationId === "poster" ? "image/jpeg" : "image/png";
      if (representationId?.includes("play") || representationId === "original") {
        return new Blob([representationId], { type: "video/mp4" });
      }
      return new Blob([representationId ?? "artifact"], { type });
    }),
    loadOlder: vi.fn(),
    open: vi.fn(),
    resolvePermission: vi.fn(),
    sendMessage: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    updateConfiguration: vi.fn(),
  };
}

function imageEditArtifact(): AgentArtifactItem {
  return {
    actions: ["image.edit"],
    artifactId: "image-1",
    filename: "result.png",
    grants: [
      {
        actions: ["image.edit"],
        grantId: "image-edit",
        representationIds: ["source"],
      },
    ],
    id: "artifact-image-1",
    kind: "artifact",
    manifest: {
      annotations: [],
      candidateRepresentationIds: ["result"],
      kind: "image_edit",
      regions: [{ height: 0.4, width: 0.3, x: 0.1, y: 0.2 }],
      resultRepresentationId: "result",
      sourceDimensions: { height: 1080, width: 1920 },
      sourceRepresentationId: "source",
    },
    mimeType: "image/png",
    representations: [
      {
        filename: "source.png",
        mediaType: "image/png",
        representationId: "source",
        revision: 11n,
        role: "source",
        status: "ready",
      },
      {
        filename: "result.png",
        mediaType: "image/png",
        representationId: "result",
        revision: 12n,
        role: "result",
        status: "ready",
      },
    ],
    revision: 12n,
    role: "image_edit",
    schemaVersion: "1",
    selectedRepresentationId: "result",
    status: "completed",
  };
}

function videoArtifact(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "video-1",
    filename: "playable.mp4",
    grants: [],
    id: "artifact-video-1",
    kind: "artifact",
    manifest: {
      derivativeRepresentationIds: ["derivative"],
      durationMillis: 65_000n,
      kind: "video",
      originalRepresentationId: "original",
      playableRepresentationId: "playable",
      posterRepresentationId: "poster",
      stage: "ready",
      thumbnailRepresentationIds: [],
    },
    mimeType: "video/mp4",
    representations: [
      videoRepresentation("playable", "playable.mp4", "video/mp4"),
      videoRepresentation("original", "original.mov", "video/quicktime"),
      videoRepresentation("poster", "poster.jpg", "image/jpeg"),
      videoRepresentation("derivative", "small.mp4", "video/mp4"),
    ],
    revision: 4n,
    role: "preview",
    schemaVersion: "1",
    selectedRepresentationId: "playable",
    status: "completed",
  };
}

function videoRepresentation(
  representationId: string,
  filename: string,
  mediaType: string,
) {
  return {
    filename,
    mediaType,
    representationId,
    revision: 4n,
    status: "ready" as const,
  };
}

function presentationArtifact(): AgentArtifactItem {
  return {
    actions: [
      "presentation.regenerate_slide",
      "presentation.select_version",
    ],
    artifactId: "deck-1",
    filename: "deck.pptx",
    grants: [
      {
        actions: [
          "presentation.regenerate_slide",
          "presentation.select_version",
        ],
        grantId: "deck-actions",
        representationIds: [],
      },
    ],
    id: "artifact-deck-1",
    kind: "artifact",
    manifest: {
      deckRevision: 8n,
      kind: "presentation",
      selectedVersionId: "v1",
      slides: [
        {
          notes: "先说明业务目标。",
          pageRepresentationId: "page-1",
          position: 1,
          slideId: "slide-1",
          thumbnailRepresentationId: "thumb-1",
          title: "项目概览",
        },
      ],
      versions: [
        { id: "v1", label: "初稿", revision: 8n },
        { id: "v2", label: "评审稿", revision: 9n },
      ],
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
      {
        filename: "slide-1.png",
        mediaType: "image/png",
        representationId: "page-1",
        revision: 8n,
        status: "ready",
      },
      {
        filename: "slide-1-thumb.png",
        mediaType: "image/png",
        representationId: "thumb-1",
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
