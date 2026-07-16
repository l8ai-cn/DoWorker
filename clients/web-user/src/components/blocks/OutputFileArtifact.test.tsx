import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { OutputFileArtifact } from "./OutputFileArtifact";

const authenticatedFetch = vi.fn();

vi.mock("@/lib/identity", () => ({
  authenticatedFetch: (...args: unknown[]) => authenticatedFetch(...args),
}));

beforeEach(() => {
  authenticatedFetch.mockResolvedValue({
    ok: true,
    blob: async () => new Blob(["artifact"]),
  });
  vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:session-file");
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
  conversationId: "session a",
  workspaceRoot: null,
  workspaceHome: null,
};

function renderArtifact(
  fileId: string,
  filename: string | null,
  contentType: string | null = null,
) {
  return render(
    <FileViewerContext.Provider value={FILE_VIEWER_CONTEXT}>
      <OutputFileArtifact fileId={fileId} filename={filename} contentType={contentType} />
    </FileViewerContext.Provider>,
  );
}

describe("OutputFileArtifact", () => {
  it("loads MP4 output through authenticated fetch and renders a blob URL", async () => {
    renderArtifact("file/1", "seedance.MP4");

    const video = await screen.findByLabelText("seedance.MP4");
    expect(video.tagName).toBe("VIDEO");
    expect(video).toHaveAttribute("controls");
    expect(video).toHaveAttribute("src", "blob:session-file");
    expect(authenticatedFetch).toHaveBeenCalledWith(
      "/v1/sessions/session%20a/resources/files/file%2F1/content",
      { signal: expect.any(AbortSignal) },
    );
  });

  it("downloads a generic file only after the user requests it", async () => {
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    renderArtifact("file 2", "storyboard.pdf");

    expect(authenticatedFetch).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Download storyboard.pdf" }));

    await waitFor(() => expect(click).toHaveBeenCalled());
    expect(authenticatedFetch).toHaveBeenCalledWith(
      "/v1/sessions/session%20a/resources/files/file%202/content",
      { signal: expect.any(AbortSignal) },
    );
  });

  it("renders an MP4 MIME response as video without relying on its filename", async () => {
    renderArtifact("file-video", "provider-output", "video/mp4");

    expect((await screen.findByLabelText("provider-output")).tagName).toBe("VIDEO");
  });

  it("uses a readable fallback when the filename is missing", () => {
    renderArtifact("file_3", null);

    expect(screen.getByText("Generated file")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Download Generated file" })).toBeInTheDocument();
  });

  it("shows an unavailable state when no session is loaded", () => {
    render(<OutputFileArtifact fileId="file_4" filename="clip.mp4" contentType="video/mp4" />);

    expect(screen.getByRole("alert")).toHaveTextContent("File unavailable");
  });

  it("shows an authenticated file load failure", async () => {
    authenticatedFetch.mockResolvedValue({ ok: false, status: 401 });
    renderArtifact("file_5", "clip.mp4");

    expect(await screen.findByRole("alert")).toHaveTextContent("Video could not be loaded");
  });

  it("shows a load failure when the MP4 element cannot load", async () => {
    renderArtifact("file_5", "clip.mp4");

    fireEvent.error(await screen.findByLabelText("clip.mp4"));

    expect(screen.getByRole("alert")).toHaveTextContent("Video could not be loaded");
  });

  it("shows a playback error when an already loaded MP4 fails", async () => {
    renderArtifact("file_6", "clip.mp4");

    const video = await screen.findByLabelText("clip.mp4");
    fireEvent.loadedData(video);
    fireEvent.error(video);

    expect(screen.getByRole("alert")).toHaveTextContent("Video playback failed");
  });

  it("revokes the blob URL on unmount", async () => {
    const view = renderArtifact("file_7", "clip.mp4");
    await screen.findByLabelText("clip.mp4");

    view.unmount();

    await waitFor(() => expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:session-file"));
  });

  it("aborts the authenticated request on unmount", async () => {
    let signal: AbortSignal | undefined;
    authenticatedFetch.mockImplementation((_path: string, init: RequestInit) => {
      signal = init.signal as AbortSignal;
      return new Promise(() => {});
    });
    const view = renderArtifact("file_8", "clip.mp4");
    await waitFor(() => expect(signal).toBeDefined());

    view.unmount();

    expect(signal?.aborted).toBe(true);
  });
});
