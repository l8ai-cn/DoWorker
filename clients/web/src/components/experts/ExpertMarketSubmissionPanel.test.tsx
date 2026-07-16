import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import {
  listExpertMarketSubmissions,
  withdrawExpertMarketRelease,
} from "@/lib/api/expertMarketApi";
import { ExpertMarketSubmissionPanel } from "./ExpertMarketSubmissionPanel";

vi.mock("@/lib/api/expertMarketApi", () => ({
  listExpertMarketSubmissions: vi.fn(),
  withdrawExpertMarketRelease: vi.fn(),
}));
vi.mock("sonner", () => ({
  toast: { error: vi.fn() },
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => ({
    pendingReview: "Pending review",
    published: "Published",
    withdrawn: "Withdrawn",
    rejected: "Rejected",
    withdraw: "Withdraw submission",
    withdrawConfirm: "Withdraw",
    withdrawFailed: "Could not unpublish this release",
    submitNew: "Submit new release",
    submitFirst: "Submit to marketplace",
    empty: "No marketplace submission",
    snapshotRequired: "Publish from a Worker before marketplace submission.",
    retry: "Retry",
  })[key] ?? key,
}));

describe("ExpertMarketSubmissionPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("withdraws the published release and refreshes its status", async () => {
    vi.mocked(listExpertMarketSubmissions)
      .mockResolvedValueOnce({
        releases: [{
          id: 18,
          application_id: 4,
          application_slug: "video-production",
          source_expert_id: 7,
          version: 2,
          status: "published",
          name: "Video Director",
          summary: "Plans video production",
          description: "",
          category: "video",
          icon: "film",
          tags: [],
          outcomes: [],
          created_at: "2026-07-15T08:00:00Z",
        }],
        total: 1,
      })
      .mockResolvedValueOnce({
        releases: [{
          id: 18,
          application_id: 4,
          application_slug: "video-production",
          source_expert_id: 7,
          version: 2,
          status: "withdrawn",
          name: "Video Director",
          summary: "Plans video production",
          description: "",
          category: "video",
          icon: "film",
          tags: [],
          outcomes: [],
          created_at: "2026-07-15T08:00:00Z",
        }],
        total: 1,
      });
    vi.mocked(withdrawExpertMarketRelease).mockResolvedValue(undefined);

    render(<ExpertMarketSubmissionPanel expertID={7} expertSlug="video-director" submissionReady />);

    expect(await screen.findByText("Published")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Withdraw submission" }));
    fireEvent.click(screen.getByRole("button", { name: "Withdraw" }));

    await waitFor(() => {
      expect(withdrawExpertMarketRelease).toHaveBeenCalledWith(18);
      expect(screen.getByText("Withdrawn")).toBeInTheDocument();
    });
  });

  it("shows the rejection reason and offers a new submission", async () => {
    vi.mocked(listExpertMarketSubmissions).mockResolvedValue({
      releases: [{
        id: 19,
        application_id: 4,
        application_slug: "video-production",
        source_expert_id: 7,
        version: 2,
        status: "rejected",
        name: "Video Director",
        summary: "Plans video production",
        description: "",
        category: "video",
        icon: "film",
        tags: [],
        outcomes: [],
        rejection_reason: "Add concrete output examples.",
        created_at: "2026-07-15T08:00:00Z",
      }],
      total: 1,
    });

    render(<ExpertMarketSubmissionPanel expertID={7} expertSlug="video-director" submissionReady />);

    expect(await screen.findByText("Rejected")).toBeInTheDocument();
    expect(screen.getByText("Add concrete output examples.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Submit new release" })).toBeEnabled();
  });

  it("reuses the existing application slug for a later expert version", async () => {
    vi.mocked(listExpertMarketSubmissions).mockResolvedValue({
      releases: [{
        id: 19,
        application_id: 4,
        application_slug: "video-production",
        source_expert_id: 7,
        version: 2,
        status: "rejected",
        name: "Video Director",
        summary: "Plans video production",
        description: "",
        category: "video",
        icon: "film",
        tags: [],
        outcomes: [],
        created_at: "2026-07-15T08:00:00Z",
      }],
      total: 1,
    });

    render(<ExpertMarketSubmissionPanel expertID={7} expertSlug="renamed-expert" submissionReady />);

    fireEvent.click(await screen.findByRole("button", { name: "Submit new release" }));

    expect(screen.getByLabelText("slugLabel"))
      .toHaveValue("video-production");
    expect(screen.getByLabelText("slugLabel")).toHaveAttribute("readonly");
  });

  it("keeps the published release visible and reports an unpublish failure", async () => {
    vi.mocked(listExpertMarketSubmissions).mockResolvedValue({
      releases: [{
        id: 18,
        application_id: 4,
        application_slug: "video-production",
        source_expert_id: 7,
        version: 2,
        status: "published",
        name: "Video Director",
        summary: "Plans video production",
        description: "",
        category: "video",
        icon: "film",
        tags: [],
        outcomes: [],
        created_at: "2026-07-15T08:00:00Z",
      }],
      total: 1,
    });
    vi.mocked(withdrawExpertMarketRelease).mockRejectedValue(
      new Error("Marketplace unavailable"),
    );

    render(<ExpertMarketSubmissionPanel expertID={7} expertSlug="video-director" submissionReady />);

    fireEvent.click(await screen.findByRole("button", { name: "Withdraw submission" }));
    fireEvent.click(screen.getByRole("button", { name: "Withdraw" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Could not unpublish this release");
      expect(screen.getByText("Published")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Withdraw" })).toBeEnabled();
    });
  });

  it("shows the empty state and recovers from a load error", async () => {
    vi.mocked(listExpertMarketSubmissions)
      .mockRejectedValueOnce(new Error("Marketplace unavailable"))
      .mockResolvedValueOnce({ releases: [], total: 0 });

    render(<ExpertMarketSubmissionPanel expertID={7} expertSlug="video-director" submissionReady />);

    expect(await screen.findByText("Marketplace unavailable")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("No marketplace submission")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Submit to marketplace" })).toBeEnabled();
  });

  it("disables submission when the expert has no Worker snapshot", async () => {
    vi.mocked(listExpertMarketSubmissions).mockResolvedValue({ releases: [], total: 0 });

    render(
      <ExpertMarketSubmissionPanel
        expertID={7}
        expertSlug="video-director"
        submissionReady={false}
      />,
    );

    expect(await screen.findByText(
      "Publish from a Worker before marketplace submission.",
    )).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Submit to marketplace" })).toBeDisabled();
  });
});
