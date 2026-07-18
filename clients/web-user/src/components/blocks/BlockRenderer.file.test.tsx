import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { authenticatedFetch } from "@/lib/identity";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { BlockRenderer } from "./BlockRenderer";

vi.mock("@/lib/identity", () => ({ authenticatedFetch: vi.fn() }));

beforeEach(() => {
  vi.mocked(authenticatedFetch).mockResolvedValue(
    new Response("video", {
      status: 200,
      headers: { "Content-Type": "video/mp4" },
    }),
  );
  vi.stubGlobal("URL", {
    ...URL,
    createObjectURL: vi.fn(() => "blob:block-file"),
    revokeObjectURL: vi.fn(),
  });
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

const FILE_VIEWER_CONTEXT = {
  openFile: () => {},
  isChangedPath: () => false,
  conversationId: "session_1",
  workspaceRoot: null,
  workspaceHome: null,
};

describe("BlockRenderer file dispatch", () => {
  it("forwards the content type to the lazy video artifact", async () => {
    render(
      <FileViewerContext.Provider value={FILE_VIEWER_CONTEXT}>
        <BlockRenderer
          items={[
            {
              kind: "file",
              itemId: "item_1",
              fileId: "file_1",
              filename: "seedance-output",
              contentType: "video/mp4",
            },
          ]}
          sessionStatus="idle"
        />
      </FileViewerContext.Provider>,
    );

    expect(authenticatedFetch).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "加载视频 seedance-output" }));
    expect(
      await screen.findByLabelText("视频预览：seedance-output"),
    ).toHaveAttribute(
      "src",
      "blob:block-file",
    );
  });
});
