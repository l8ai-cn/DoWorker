import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const mockList = vi.fn();
vi.mock("@/lib/api/admin", () => ({
  listExpertMarketReleases: (...args: unknown[]) => mockList(...args),
  getExpertMarketRelease: vi.fn(),
  approveExpertMarketRelease: vi.fn(),
  rejectExpertMarketRelease: vi.fn(),
}));

import ExpertMarketPage from "../page";

const videoRelease = {
  id: 12,
  name: "Video Expert",
  summary: "Build production videos",
  category: "media",
  version: 3,
  status: "pending",
  submitted_at: "2026-07-14T08:00:00Z",
};

const audioRelease = {
  ...videoRelease,
  id: 13,
  name: "Audio Expert",
  summary: "Build production audio",
};

describe("ExpertMarketPage pagination", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses response pagination metadata for next and previous requests", async () => {
    mockList
      .mockResolvedValueOnce(page(videoRelease, 120, 0))
      .mockResolvedValueOnce(page(audioRelease, 120, 50))
      .mockResolvedValueOnce(page(videoRelease, 120, 0));

    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    expect(screen.getByText("第 1 / 3 页")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "上一页" })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    await screen.findByText("Audio Expert");
    expect(mockList).toHaveBeenLastCalledWith({
      status: "pending",
      limit: 50,
      offset: 50,
    });
    expect(screen.getByText("第 2 / 3 页")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "上一页" }));
    await waitFor(() => {
      expect(mockList).toHaveBeenLastCalledWith({
        status: "pending",
        limit: 50,
        offset: 0,
      });
    });
  });

  it("resets the offset when the status filter changes", async () => {
    mockList
      .mockResolvedValueOnce(page(videoRelease, 120, 0))
      .mockResolvedValueOnce(page(audioRelease, 120, 50))
      .mockResolvedValueOnce(page(videoRelease, 1, 0));

    render(<ExpertMarketPage />);
    await screen.findByText("Video Expert");
    fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    await screen.findByText("Audio Expert");
    fireEvent.click(screen.getByRole("button", { name: "已发布" }));

    await waitFor(() => {
      expect(mockList).toHaveBeenLastCalledWith({
        status: "published",
        limit: 50,
        offset: 0,
      });
    });
  });
});

function page(item: typeof videoRelease, total: number, offset: number) {
  return { items: [item], total, limit: 50, offset };
}
