import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { authenticatedFetch } from "@/lib/identity";
import { FileViewerContext } from "@/shell/FileViewerContext";
import { OutputFileArtifact } from "./OutputFileArtifact";

vi.mock("@/lib/identity", () => ({ authenticatedFetch: vi.fn() }));

const fetchMock = vi.mocked(authenticatedFetch);
const createObjectUrl = vi.fn(() => "blob:session-file");
const revokeObjectUrl = vi.fn();
const anchorClick = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});

beforeEach(() => {
  fetchMock.mockResolvedValue(
    new Response(new Blob(["file"], { type: "application/pdf" }), { status: 200 }),
  );
  vi.stubGlobal("URL", {
    ...URL,
    createObjectURL: createObjectUrl,
    revokeObjectURL: revokeObjectUrl,
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
  it("does not download an artifact before the user requests it", () => {
    renderArtifact("file/1", "seedance.mp4", "video/mp4");

    expect(fetchMock).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "加载视频 seedance.mp4" })).toBeInTheDocument();
  });

  it("loads video through the unified authenticated request", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response("video", {
        status: 200,
        headers: { "Content-Type": "video/mp4" },
      }),
    );
    renderArtifact("file/1", "seedance.mp4", "video/mp4");

    fireEvent.click(screen.getByRole("button", { name: "加载视频 seedance.mp4" }));

    expect(await screen.findByLabelText("seedance.mp4")).toHaveAttribute(
      "src",
      "blob:session-file",
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/v1/sessions/session%20a/resources/files/file%2F1/content",
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });

  it("uses the response MIME type when event metadata and filename are ambiguous", async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: new Headers({ "Content-Type": "video/mp4" }),
      blob: async () => new Blob(["video"]),
    } as Response);
    renderArtifact("file_2", "seedance-output");

    fireEvent.click(screen.getByRole("button", { name: "下载 seedance-output" }));

    expect(await screen.findByLabelText("seedance-output")).toHaveAttribute(
      "src",
      "blob:session-file",
    );
    expect(anchorClick).not.toHaveBeenCalled();
  });

  it("downloads a generic artifact only after the user requests it", async () => {
    renderArtifact("file 3", "storyboard.pdf");

    fireEvent.click(screen.getByRole("button", { name: "下载 storyboard.pdf" }));

    await waitFor(() => expect(anchorClick).toHaveBeenCalledTimes(1));
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("retries a failed generic artifact download", async () => {
    fetchMock
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockResolvedValueOnce(
        new Response(new Blob(["file"], { type: "application/pdf" }), { status: 200 }),
      );
    renderArtifact("file_retry", "storyboard.pdf");

    fireEvent.click(screen.getByRole("button", { name: "下载 storyboard.pdf" }));
    fireEvent.click(await screen.findByRole("button", { name: "重试" }));

    await waitFor(() => expect(anchorClick).toHaveBeenCalledTimes(1));
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("uses a readable fallback when the filename is missing", () => {
    renderArtifact("file_4", null);

    expect(screen.getByText("生成文件")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "下载 生成文件" })).toBeInTheDocument();
  });

  it("shows an unavailable state when no session is loaded", () => {
    render(<OutputFileArtifact fileId="file_5" filename="clip.mp4" contentType="video/mp4" />);

    expect(screen.getByRole("alert")).toHaveTextContent("文件不可用");
  });

  it("shows a request failure after an authenticated request fails", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 401 }));
    renderArtifact("file_6", "clip.mp4", "video/mp4");

    fireEvent.click(screen.getByRole("button", { name: "加载视频 clip.mp4" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("视频加载失败");
  });

  it("shows a playback failure after video metadata loaded", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(new Blob(["video"], { type: "video/mp4" }), { status: 200 }),
    );
    renderArtifact("file_7", "clip.mp4", "video/mp4");
    fireEvent.click(screen.getByRole("button", { name: "加载视频 clip.mp4" }));
    const video = await screen.findByLabelText("clip.mp4");

    fireEvent.loadedData(video);
    fireEvent.error(video);

    expect(screen.getByRole("alert")).toHaveTextContent("视频播放失败");
  });

  it("aborts a pending request when the artifact unmounts", async () => {
    let requestSignal: AbortSignal | undefined;
    fetchMock.mockImplementationOnce((_path, init) => {
      requestSignal = init?.signal ?? undefined;
      return new Promise<Response>(() => {});
    });
    const view = renderArtifact("file_8", "clip.mp4", "video/mp4");
    fireEvent.click(screen.getByRole("button", { name: "加载视频 clip.mp4" }));
    await waitFor(() => expect(requestSignal).toBeDefined());

    view.unmount();

    expect(requestSignal?.aborted).toBe(true);
  });

  it("revokes the object URL on unmount", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(new Blob(["video"], { type: "video/mp4" }), { status: 200 }),
    );
    const view = renderArtifact("file_9", "clip.mp4", "video/mp4");
    fireEvent.click(screen.getByRole("button", { name: "加载视频 clip.mp4" }));
    await screen.findByLabelText("clip.mp4");

    view.unmount();

    expect(revokeObjectUrl).toHaveBeenCalledWith("blob:session-file");
  });
});
