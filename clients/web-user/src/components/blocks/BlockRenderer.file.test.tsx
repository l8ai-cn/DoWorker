import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { BlockRenderer } from "./BlockRenderer";

afterEach(cleanup);

const FILE_VIEWER_CONTEXT = {
  openFile: () => {},
  isChangedPath: () => false,
  conversationId: "session_1",
  workspaceRoot: null,
  workspaceHome: null,
};

describe("BlockRenderer file dispatch", () => {
  it("renders a file item through OutputFileArtifact", () => {
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

    expect(screen.getByLabelText("seedance.mp4")).toHaveAttribute(
      "src",
      "/v1/sessions/session_1/resources/files/file_1/content",
    );
  });
});
