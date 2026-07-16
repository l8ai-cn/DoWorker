import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { OutputFileArtifact } from "./OutputFileArtifact";

afterEach(cleanup);

const FILE_VIEWER_CONTEXT = {
  openFile: () => {},
  isChangedPath: () => false,
  conversationId: "session a",
  workspaceRoot: null,
  workspaceHome: null,
};

function renderArtifact(fileId: string, filename: string | null) {
  return render(
    <FileViewerContext.Provider value={FILE_VIEWER_CONTEXT}>
      <OutputFileArtifact fileId={fileId} filename={filename} />
    </FileViewerContext.Provider>,
  );
}

describe("OutputFileArtifact", () => {
  it("renders MP4 output with a same-origin session file URL", () => {
    renderArtifact("file/1", "seedance.MP4");

    const video = screen.getByLabelText("seedance.MP4");
    expect(video.tagName).toBe("VIDEO");
    expect(video).toHaveAttribute("controls");
    expect(video).toHaveAttribute(
      "src",
      "/v1/sessions/session%20a/resources/files/file%2F1/content",
    );
  });

  it("renders a generic file as a same-origin download link", () => {
    renderArtifact("file 2", "storyboard.pdf");

    const link = screen.getByRole("link", { name: "Download storyboard.pdf" });
    expect(link).toHaveAttribute(
      "href",
      "/v1/sessions/session%20a/resources/files/file%202/content",
    );
    expect(link).toHaveAttribute("download", "storyboard.pdf");
  });

  it("uses a readable fallback when the filename is missing", () => {
    renderArtifact("file_3", null);

    expect(screen.getByText("Generated file")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Download Generated file" })).toHaveAttribute(
      "download",
      "generated-file",
    );
  });

  it("shows an unavailable state when no session is loaded", () => {
    render(<OutputFileArtifact fileId="file_4" filename="clip.mp4" />);

    expect(screen.getByRole("alert")).toHaveTextContent("File unavailable");
  });

  it("shows a load failure when the MP4 cannot load", () => {
    renderArtifact("file_5", "clip.mp4");

    fireEvent.error(screen.getByLabelText("clip.mp4"));

    expect(screen.getByRole("alert")).toHaveTextContent("Video could not be loaded");
  });

  it("shows a playback error when an already loaded MP4 fails", () => {
    renderArtifact("file_6", "clip.mp4");

    const video = screen.getByLabelText("clip.mp4");
    fireEvent.loadedData(video);
    fireEvent.error(video);

    expect(screen.getByRole("alert")).toHaveTextContent("Video playback failed");
  });
});
