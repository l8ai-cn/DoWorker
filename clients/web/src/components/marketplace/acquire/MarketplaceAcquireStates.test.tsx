import { describe, expect, it } from "vitest";
import { render, screen } from "@/test/test-utils";

import { SuccessState } from "./MarketplaceAcquireStates";

describe("MarketplaceAcquireStates", () => {
  it("sends a successful acquisition to its application instance", () => {
    const props = {
      organization: { id: 9, slug: "dev-org", name: "研发组织" },
      installationID: "installation-1",
    };
    render(
      <SuccessState {...props} />,
    );

    expect(screen.getByRole("link", { name: "去应用中心开始第一个任务" }))
      .toHaveAttribute("href", "/dev-org/applications/installation-1");
  });
});
