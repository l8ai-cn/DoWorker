import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { VideoPreviewDelivery } from "../VideoPreviewDelivery";

const mocks = vi.hoisted(() => ({
  getSession: vi.fn(),
}));

vi.mock("@/lib/api/podPreview", () => ({
  getPodPreviewSession: mocks.getSession,
}));

const t = (key: string) => key;

describe("VideoPreviewDelivery", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.getSession.mockResolvedValue({
      preview_base_url: "https://preview.l8ai.cn/preview/video-worker-1/",
      session_url:
        "https://preview.l8ai.cn/preview/video-worker-1/__session?token=test",
      expires_at: "2026-07-18T12:00:00Z",
    });
  });

  it("embeds the authenticated pod preview session", async () => {
    render(
      <VideoPreviewDelivery
        orgSlug="dev-org"
        podKey="video-worker-1"
        t={t}
      />,
    );

    const frame = await screen.findByTitle(
      "videoWorker.videoPreview",
    );
    expect(frame).toHaveAttribute(
      "src",
      "https://preview.l8ai.cn/preview/video-worker-1/__session?token=test",
    );
    expect(mocks.getSession).toHaveBeenCalledWith(
      "dev-org",
      "video-worker-1",
    );
  });
});
