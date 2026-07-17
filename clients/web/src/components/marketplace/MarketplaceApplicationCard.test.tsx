import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MarketplaceApplicationCard } from "./MarketplaceApplicationCard";

vi.mock("./MarketplaceInstallButton", () => ({
  MarketplaceInstallButton: () => <button type="button">安装</button>,
}));

describe("MarketplaceApplicationCard", () => {
  it("renders the video production icon contract", () => {
    const { container } = render(
      <MarketplaceApplicationCard
        application={{
          slug: "video-production-expert",
          name: "视频制作专家",
          summary: "制作短视频",
          description: "从脚本到成片",
          category: "video",
          icon: "clapperboard",
          agent_slug: "video-studio",
          skill_slugs: ["remotion-best-practices"],
          tags: ["short-video"],
          outcomes: ["playable mp4"],
          version: 1,
          featured: true,
        }}
      />,
    );

    expect(screen.getByText("视频制作专家")).toBeInTheDocument();
    expect(container.querySelector("svg.lucide-clapperboard")).not.toBeNull();
  });
});
