import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { WorkerTabContent } from "../WorkerTabContent";

const mocks = vi.hoisted(() => ({
  getWorkerContext: vi.fn(),
  sendPrompt: vi.fn(),
  updatePreview: vi.fn(),
  upsertPod: vi.fn(),
}));

vi.mock("@/lib/api/podWorkerContext", () => ({
  getPodWorkerContext: mocks.getWorkerContext,
}));

vi.mock("@/lib/api/facade/podConnect", () => ({
  sendPodPrompt: mocks.sendPrompt,
  updatePodPreviewConfig: mocks.updatePreview,
}));

vi.mock("@/stores/pod", () => ({
  usePodStore: { getState: () => ({ upsertPod: mocks.upsertPod }) },
}));

const t = (key: string, params?: Record<string, string | number>) =>
  params?.port ? `${key}:${params.port}` : key;

const pod = {
  pod_key: "video-worker-1",
  status: "running" as const,
  alias: "video-production-expert",
  agent: { slug: "video-studio", name: "Video Studio" },
  worker_spec_snapshot_id: 91,
};

describe("WorkerTabContent", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.getWorkerContext.mockResolvedValue({
      snapshot_id: 91,
      alias: "video-production-expert",
      expert: {
        id: 7,
        name: "视频制作专家",
        slug: "video-production-expert",
      },
      skill_slugs: [
        "short-video-directing",
        "remotion-video-production",
        "video-delivery-qa",
      ],
    });
    mocks.updatePreview.mockResolvedValue({
      ...pod,
      preview_port: 4173,
      preview_path: "/oilan-video/index.html",
    });
    mocks.sendPrompt.mockResolvedValue("sent");
  });

  it("shows the linked expert and its skills", async () => {
    render(
      <WorkerTabContent
        selectedPodKey={pod.pod_key}
        pod={pod}
        orgSlug="dev-org"
        t={t}
      />,
    );

    expect(await screen.findByText("视频制作专家")).toBeInTheDocument();
    expect(screen.getByText("short-video-directing")).toBeInTheDocument();
    expect(screen.getByText("remotion-video-production")).toBeInTheDocument();
    expect(screen.getByText("video-delivery-qa")).toBeInTheDocument();
  });

  it("enables preview and sends a video delivery task", async () => {
    render(
      <WorkerTabContent
        selectedPodKey={pod.pod_key}
        pod={pod}
        orgSlug="dev-org"
        t={t}
      />,
    );

    fireEvent.change(
      screen.getByPlaceholderText("videoWorker.taskPlaceholder"),
      { target: { value: "生成一条18秒OILAN竖屏短视频" } },
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "videoWorker.sendTask",
      }),
    );

    await waitFor(() => {
      expect(mocks.updatePreview).toHaveBeenCalledWith(
        "dev-org",
        pod.pod_key,
        4173,
        "/oilan-video/index.html",
      );
      expect(mocks.sendPrompt).toHaveBeenCalledTimes(1);
    });
    expect(mocks.upsertPod).toHaveBeenCalledWith(
      expect.objectContaining({ preview_port: 4173 }),
    );
    expect(mocks.sendPrompt.mock.calls[0][2]).toContain(
      "生成一条18秒OILAN竖屏短视频",
    );
    expect(mocks.sendPrompt.mock.calls[0][2]).toContain(
      "delivery/oilan-video-preview.mp4",
    );
    expect(mocks.sendPrompt.mock.invocationCallOrder[0]).toBeLessThan(
      mocks.updatePreview.mock.invocationCallOrder[0],
    );
    expect(mocks.sendPrompt.mock.calls[0][2]).toContain(
      "python3 -m http.server 4173",
    );
  });
});
