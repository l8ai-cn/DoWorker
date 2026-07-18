import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { FileContentResponse } from "@/hooks/useFileContent";
import { isWorkspaceVideoFile, WorkspaceVideoViewer } from "./WorkspaceVideoViewer";

const VIDEO_DATA: FileContentResponse = {
  object: "session.environment.filesystem.file_content",
  path: "output/clip.mp4",
  content_type: "video/mp4",
  encoding: "base64",
  content: "AAAA",
  bytes: 3,
};

beforeEach(() => {
  vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:workspace-video");
  vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("isWorkspaceVideoFile", () => {
  it("prefers a video MIME type and falls back to supported extensions", () => {
    expect(isWorkspaceVideoFile("output/render.bin", "video/mp4")).toBe(true);
    expect(isWorkspaceVideoFile("output/render.MP4", null)).toBe(true);
    expect(isWorkspaceVideoFile("output/render.webm", null)).toBe(true);
    expect(isWorkspaceVideoFile("output/render.mp4", "application/octet-stream")).toBe(false);
  });
});

describe("WorkspaceVideoViewer", () => {
  it("renders a workspace MP4 through a blob URL", async () => {
    render(<WorkspaceVideoViewer data={VIDEO_DATA} path={VIDEO_DATA.path} />);

    const video = await screen.findByLabelText("clip.mp4");
    expect(video).toHaveAttribute("controls");
    expect(video).toHaveAttribute("src", "blob:workspace-video");
  });

  it("shows a playback error when the video element fails", async () => {
    render(<WorkspaceVideoViewer data={VIDEO_DATA} path={VIDEO_DATA.path} />);

    fireEvent.error(await screen.findByLabelText("clip.mp4"));

    expect(screen.getByRole("alert")).toHaveTextContent("Video playback failed");
  });

  it("rejects truncated video bytes before creating an object URL", () => {
    render(
      <WorkspaceVideoViewer data={{ ...VIDEO_DATA, truncated: true }} path={VIDEO_DATA.path} />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("too large to preview");
    expect(URL.createObjectURL).not.toHaveBeenCalled();
  });

  it("revokes its object URL on unmount", async () => {
    const view = render(<WorkspaceVideoViewer data={VIDEO_DATA} path={VIDEO_DATA.path} />);
    await screen.findByLabelText("clip.mp4");

    view.unmount();

    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:workspace-video");
  });
});
