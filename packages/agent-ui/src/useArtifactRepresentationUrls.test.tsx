import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";
import { useArtifactRepresentationUrls } from "./useArtifactRepresentationUrls";

describe("useArtifactRepresentationUrls", () => {
  it("does not reload an unchanged representation when the item object is rebuilt", async () => {
    const loadArtifact = vi.fn(async () => {
      return new Blob(["image"], { type: "image/png" });
    });
    const createObjectURL = vi
      .spyOn(URL, "createObjectURL")
      .mockReturnValue("blob:artifact");
    const revokeObjectURL = vi
      .spyOn(URL, "revokeObjectURL")
      .mockImplementation(() => undefined);
    const runtime = { loadArtifact } as unknown as AgentSessionRuntime;
    const item = artifactItem();
    const { rerender, result, unmount } = renderHook(
      ({ value }) =>
        useArtifactRepresentationUrls(
          value,
          runtime,
          "session-1",
          ["source"],
        ),
      { initialProps: { value: item } },
    );

    await waitFor(() => expect(result.current.source?.status).toBe("ready"));
    rerender({
      value: {
        ...item,
        representations: item.representations.map((representation) => ({
          ...representation,
        })),
      },
    });

    expect(loadArtifact).toHaveBeenCalledTimes(1);
    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:artifact");
    createObjectURL.mockRestore();
    revokeObjectURL.mockRestore();
  });
});

function artifactItem(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "artifact-1",
    filename: "source.png",
    grants: [{
      actions: ["artifact.download"],
      grantId: "grant-download",
      representationIds: [],
    }],
    id: "artifact-item-1",
    kind: "artifact",
    manifest: null,
    mimeType: "image/png",
    representations: [
      {
        mediaType: "image/png",
        representationId: "source",
        revision: 1n,
        status: "ready",
      },
    ],
    revision: 1n,
    role: "image",
    schemaVersion: "1",
    selectedRepresentationId: "source",
    status: "completed",
  };
}
