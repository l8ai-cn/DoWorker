import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockToastSuccess = vi.fn();
const mockToastError = vi.fn();
vi.mock("sonner", () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}));

const mockList = vi.fn();
const mockGet = vi.fn();
const mockApprove = vi.fn();
const mockReject = vi.fn();
vi.mock("@/lib/api/admin", () => ({
  listExpertMarketReleases: (...args: unknown[]) => mockList(...args),
  getExpertMarketRelease: (...args: unknown[]) => mockGet(...args),
  approveExpertMarketRelease: (...args: unknown[]) => mockApprove(...args),
  rejectExpertMarketRelease: (...args: unknown[]) => mockReject(...args),
}));

import ExpertMarketPage from "../page";

const release = {
  id: 12,
  application_id: 22,
  source_expert_id: 32,
  publisher_organization_id: 42,
  publisher_user_id: 52,
  version: 3,
  status: "pending",
  name: "Video Expert",
  summary: "Build production videos",
  description: "Detailed description",
  category: "media",
  icon: "video",
  tags: ["video", "editing"],
  outcomes: ["Published video"],
  featured: false,
  expert_snapshot_json: "{\"model\":\"codex\"}",
  worker_spec_snapshot_json: "{\"runtime\":\"runner\"}",
  skill_dependencies_json: "[{\"skill_id\":7,\"slug\":\"remotion\",\"version\":1}]",
  submitted_at: "2026-07-14T08:00:00Z",
  created_at: "2026-07-14T07:00:00Z",
};

const anotherRelease = {
  ...release,
  id: 13,
  application_id: 23,
  name: "Audio Expert",
  summary: "Build production audio",
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

describe("ExpertMarketPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockList.mockResolvedValue({ items: [release], total: 1, limit: 50, offset: 0 });
    mockGet.mockResolvedValue(release);
    mockApprove.mockResolvedValue({ ...release, status: "published" });
    mockReject.mockResolvedValue({ ...release, status: "rejected" });
  });

  it("loads pending releases and switches all supported status filters", async () => {
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    expect(mockList).toHaveBeenCalledWith({ status: "pending", limit: 50, offset: 0 });

    for (const [label, status] of [
      ["已发布", "published"],
      ["已驳回", "rejected"],
      ["已撤回", "withdrawn"],
    ]) {
      fireEvent.click(screen.getByRole("button", { name: label }));
      await waitFor(() => {
        expect(mockList).toHaveBeenLastCalledWith({ status, limit: 50, offset: 0 });
      });
    }
  });

  it("ignores an older list response after the status filter changes", async () => {
    const pendingRequest = deferred<{
      items: typeof release[];
      total: number;
      limit: number;
      offset: number;
    }>();
    const publishedRequest = deferred<{
      items: typeof release[];
      total: number;
      limit: number;
      offset: number;
    }>();
    mockList
      .mockReturnValueOnce(pendingRequest.promise)
      .mockReturnValueOnce(publishedRequest.promise);

    render(<ExpertMarketPage />);
    fireEvent.click(screen.getByRole("button", { name: "已发布" }));
    publishedRequest.resolve({
      items: [anotherRelease],
      total: 1,
      limit: 50,
      offset: 0,
    });
    await screen.findByText("Audio Expert");

    await act(async () => {
      pendingRequest.resolve({
        items: [release],
        total: 1,
        limit: 50,
        offset: 0,
      });
    });
    expect(screen.queryByText("Video Expert")).not.toBeInTheDocument();
  });

  it("renders loading, empty, and retryable error states", async () => {
    mockList.mockReturnValueOnce(new Promise(() => {}));
    const { unmount } = render(<ExpertMarketPage />);
    expect(screen.getByText("正在加载审核记录...")).toBeInTheDocument();
    unmount();

    mockList.mockResolvedValueOnce({ items: [], total: 0, limit: 50, offset: 0 });
    const emptyView = render(<ExpertMarketPage />);
    await screen.findByText("当前状态下暂无发布记录");
    emptyView.unmount();

    mockList.mockRejectedValueOnce(new Error("network unavailable"));
    render(<ExpertMarketPage />);
    await screen.findByText("审核记录加载失败");
    const callsBeforeRetry = mockList.mock.calls.length;
    fireEvent.click(screen.getByRole("button", { name: "重新加载" }));
    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(callsBeforeRetry + 1);
    });
  });

  it("shows detail snapshots and skill dependencies", async () => {
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    fireEvent.click(screen.getByRole("button", { name: "查看详情" }));

    await screen.findByText("专家快照");
    expect(mockGet).toHaveBeenCalledWith(12);
    expect(screen.getByText(/"model": "codex"/)).toBeInTheDocument();
    expect(screen.getByText(/"runtime": "runner"/)).toBeInTheDocument();
    expect(screen.getByText("remotion")).toBeInTheDocument();
    expect(screen.getByText("版本 1")).toBeInTheDocument();
  });

  it("keeps the last clicked release when detail responses finish out of order", async () => {
    const firstDetail = deferred<typeof release>();
    const secondDetail = deferred<typeof anotherRelease>();
    mockList.mockResolvedValueOnce({
      items: [release, anotherRelease],
      total: 2,
      limit: 50,
      offset: 0,
    });
    mockGet
      .mockReturnValueOnce(firstDetail.promise)
      .mockReturnValueOnce(secondDetail.promise);

    render(<ExpertMarketPage />);
    await screen.findByText("Audio Expert");
    const viewButtons = screen.getAllByRole("button", { name: "查看详情" });
    fireEvent.click(viewButtons[0]);
    fireEvent.click(viewButtons[1]);

    secondDetail.resolve(anotherRelease);
    await screen.findByRole("heading", { name: "Audio Expert" });
    await act(async () => {
      firstDetail.resolve(release);
    });
    expect(
      screen.queryByRole("heading", { name: "Video Expert" }),
    ).not.toBeInTheDocument();
  });

  it("approves and refreshes the current list", async () => {
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    fireEvent.click(screen.getByRole("button", { name: "查看详情" }));
    await screen.findByRole("button", { name: "批准发布" });
    fireEvent.click(screen.getByRole("button", { name: "批准发布" }));

    await waitFor(() => expect(mockApprove).toHaveBeenCalledWith(12));
    expect(mockToastSuccess).toHaveBeenCalledWith("专家发布已批准");
    expect(mockList).toHaveBeenCalledTimes(2);
  });

  it("does not replace a newly selected release with an older approval response", async () => {
    const approval = deferred<typeof release>();
    mockList.mockResolvedValue({
      items: [release, anotherRelease],
      total: 2,
      limit: 50,
      offset: 0,
    });
    mockGet.mockImplementation(async (releaseId: number) =>
      releaseId === release.id ? release : anotherRelease,
    );
    mockApprove.mockReturnValueOnce(approval.promise);

    render(<ExpertMarketPage />);
    await screen.findByText("Audio Expert");
    const viewButtons = screen.getAllByRole("button", { name: "查看详情" });
    fireEvent.click(viewButtons[0]);
    await screen.findByRole("heading", { name: "Video Expert" });
    fireEvent.click(screen.getByRole("button", { name: "批准发布" }));
    fireEvent.click(viewButtons[1]);
    await screen.findByRole("heading", { name: "Audio Expert" });

    await act(async () => {
      approval.resolve({ ...release, status: "published" });
    });
    expect(
      screen.queryByRole("heading", { name: "Video Expert" }),
    ).not.toBeInTheDocument();
    expect(mockApprove).toHaveBeenCalledWith(release.id);
  });

  it("requires a rejection reason and reports action errors", async () => {
    mockReject.mockRejectedValueOnce(new Error("review conflict"));
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    fireEvent.click(screen.getByRole("button", { name: "查看详情" }));
    await screen.findByRole("button", { name: "驳回" });
    fireEvent.click(screen.getByRole("button", { name: "驳回" }));
    fireEvent.click(screen.getByRole("button", { name: "确认驳回" }));
    expect(screen.getByText("请输入驳回理由")).toBeInTheDocument();
    expect(mockReject).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText("驳回理由"), {
      target: { value: "Missing license" },
    });
    fireEvent.click(screen.getByRole("button", { name: "确认驳回" }));

    await waitFor(() => {
      expect(mockReject).toHaveBeenCalledWith(12, "Missing license");
      expect(mockToastError).toHaveBeenCalledWith("review conflict");
    });
  });

  it("does not replace a newly selected release with an older rejection response", async () => {
    const rejection = deferred<typeof release>();
    mockList.mockResolvedValue({
      items: [release, anotherRelease],
      total: 2,
      limit: 50,
      offset: 0,
    });
    mockGet.mockImplementation(async (releaseId: number) =>
      releaseId === release.id ? release : anotherRelease,
    );
    mockReject.mockReturnValueOnce(rejection.promise);

    render(<ExpertMarketPage />);
    await screen.findByText("Audio Expert");
    const viewButtons = screen.getAllByRole("button", { name: "查看详情" });
    fireEvent.click(viewButtons[0]);
    await screen.findByRole("heading", { name: "Video Expert" });
    fireEvent.click(screen.getByRole("button", { name: "驳回" }));
    fireEvent.change(screen.getByLabelText("驳回理由"), {
      target: { value: "Missing license" },
    });
    fireEvent.click(screen.getByRole("button", { name: "确认驳回" }));
    fireEvent.click(viewButtons[1]);
    await screen.findByRole("heading", { name: "Audio Expert" });

    await act(async () => {
      rejection.resolve({ ...release, status: "rejected" });
    });
    expect(
      screen.queryByRole("heading", { name: "Video Expert" }),
    ).not.toBeInTheDocument();
    expect(mockReject).toHaveBeenCalledWith(release.id, "Missing license");
  });

  it("refreshes on demand and shows feedback", async () => {
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    fireEvent.click(screen.getByRole("button", { name: "刷新" }));
    await waitFor(() => expect(mockList).toHaveBeenCalledTimes(2));
    expect(mockToastSuccess).toHaveBeenCalledWith("审核列表已刷新");
  });

  it("reports a manual refresh failure without a success message", async () => {
    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    mockList.mockRejectedValueOnce(new Error("refresh unavailable"));
    fireEvent.click(screen.getByRole("button", { name: "刷新" }));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("refresh unavailable");
    });
    expect(mockToastSuccess).not.toHaveBeenCalledWith("审核列表已刷新");
  });
});
