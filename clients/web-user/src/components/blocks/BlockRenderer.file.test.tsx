import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { BlockRenderer } from "./BlockRenderer";

const authenticatedFetch = vi.fn();

vi.mock("@/lib/identity", () => ({
  authenticatedFetch: (...args: unknown[]) => authenticatedFetch(...args),
}));

beforeEach(() => {
  authenticatedFetch.mockResolvedValue({
    ok: true,
    blob: async () => new Blob(["video"]),
  });
  vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:seedance-video");
  vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  authenticatedFetch.mockReset();
});

const FILE_VIEWER_CONTEXT = {
  openFile: () => {},
  isChangedPath: () => false,
  conversationId: "session_1",
  workspaceRoot: null,
  workspaceHome: null,
};

describe("BlockRenderer file dispatch", () => {
  it("renders a file item through OutputFileArtifact", async () => {
    render(
      <FileViewerContext.Provider value={FILE_VIEWER_CONTEXT}>
        <BlockRenderer
          items={[
            {
              kind: "file",
              itemId: "item_1",
              fileId: "file_1",
              filename: "seedance.mp4",
              contentType: "video/mp4",
            },
          ]}
          sessionStatus="idle"
        />
      </FileViewerContext.Provider>,
    );

    expect(await screen.findByLabelText("seedance.mp4")).toHaveAttribute(
      "src",
      "blob:seedance-video",
    );
    expect(authenticatedFetch).toHaveBeenCalledWith(
      "/v1/sessions/session_1/resources/files/file_1/content",
      { signal: expect.any(AbortSignal) },
    );
  });
});
