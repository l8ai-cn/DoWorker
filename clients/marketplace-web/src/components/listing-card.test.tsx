import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ListingCard } from "./listing-card";

describe("ListingCard", () => {
  it("shows discovery and trust information without inventing missing API data", () => {
    render(
      <ListingCard
        listing={{
          listing_id: "1",
          listing_version_id: "11",
          slug: "software-delivery-expert",
          resource_type: "application",
          display_name: "软件交付专家",
          tagline: "完成可验证的软件交付",
          publisher: {
            slug: "do-worker",
            display_name: "Do Worker",
            verified: true,
          },
          spaces: [{ slug: "software-delivery", name: "软件交付" }],
          quota: {
            mode: "per_install",
            estimated_credits_micro: "20000000",
          },
          published_at: "2026-07-11T08:00:00Z",
        }}
      />,
    );

    expect(screen.getByText("应用")).toBeInTheDocument();
    expect(screen.getByText("已认证")).toBeInTheDocument();
    expect(screen.getByText("启用需 20 市场额度")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看软件交付专家" })).toHaveAttribute(
      "href",
      "/listings/software-delivery-expert",
    );
  });
});
