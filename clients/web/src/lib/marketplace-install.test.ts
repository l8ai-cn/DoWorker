import { describe, expect, it, vi } from "vitest";

import { lightFetch } from "@/lib/light-auth";
import { installMarketplaceApplication } from "./marketplace-install";

vi.mock("@/lib/light-auth", () => ({
  lightFetch: vi.fn(),
}));

describe("installMarketplaceApplication", () => {
  it("sends the explicitly selected model resource", async () => {
    vi.mocked(lightFetch).mockResolvedValue({
      expert: { slug: "video-production-expert" },
      already_installed: false,
    });

    await installMarketplaceApplication(
      "acme",
      "video-production-expert",
      42,
    );

    expect(lightFetch).toHaveBeenCalledWith(
      "/api/v1/orgs/acme/marketplace/experts/video-production-expert/install",
      {
        method: "POST",
        authenticated: true,
        body: { model_resource_id: 42 },
      },
    );
  });
});
