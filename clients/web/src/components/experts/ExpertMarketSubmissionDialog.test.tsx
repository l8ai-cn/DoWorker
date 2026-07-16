import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { submitExpertMarketRelease } from "@/lib/api/expertMarketApi";
import { ExpertMarketSubmissionDialog } from "./ExpertMarketSubmissionDialog";

vi.mock("@/lib/api/expertMarketApi", () => ({
  submitExpertMarketRelease: vi.fn(),
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => ({
    dialogTitle: "Submit expert to marketplace",
    slugLabel: "Marketplace slug",
    summaryLabel: "Summary",
    categoryLabel: "Category",
    tagsLabel: "Tags",
    outcomesLabel: "Outcomes",
    submit: "Submit for review",
  })[key] ?? key,
}));

describe("ExpertMarketSubmissionDialog", () => {
  it("exposes an accessible dialog title", () => {
    render(
      <ExpertMarketSubmissionDialog
        expertSlug="video-director"
        marketSlug="video-production"
        open
        onOpenChange={vi.fn()}
        onSubmitted={vi.fn()}
      />,
    );

    expect(screen.getByRole("dialog", {
      name: "Submit expert to marketplace",
    })).toBeInTheDocument();
  });

  it("submits required market metadata and normalizes list fields", async () => {
    vi.mocked(submitExpertMarketRelease).mockResolvedValue(undefined);
    const onSubmitted = vi.fn();

    render(
      <ExpertMarketSubmissionDialog
        expertSlug="video-director"
        marketSlug="video-production"
        open
        onOpenChange={vi.fn()}
        onSubmitted={onSubmitted}
      />,
    );

    const submit = screen.getByRole("button", { name: "Submit for review" });
    expect(submit).toBeDisabled();
    expect(screen.getByLabelText("Marketplace slug")).toHaveValue("video-production");
    fireEvent.change(screen.getByLabelText("Summary"), {
      target: { value: "Plans production-ready videos" },
    });
    fireEvent.change(screen.getByLabelText("Category"), {
      target: { value: "video" },
    });
    fireEvent.change(screen.getByLabelText("Tags"), {
      target: { value: "video, production, video" },
    });
    fireEvent.change(screen.getByLabelText("Outcomes"), {
      target: { value: "shot list\nproduction plan" },
    });
    expect(submit).toBeEnabled();
    fireEvent.click(submit);

    await waitFor(() => {
      expect(submitExpertMarketRelease).toHaveBeenCalledWith("video-director", {
        slug: "video-production",
        summary: "Plans production-ready videos",
        description: "",
        category: "video",
        icon: "rocket",
        tags: ["video", "production"],
        outcomes: ["shot list", "production plan"],
      });
      expect(onSubmitted).toHaveBeenCalledTimes(1);
    });
  });
});
