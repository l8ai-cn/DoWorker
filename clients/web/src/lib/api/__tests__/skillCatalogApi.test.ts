import { beforeEach, describe, expect, it, vi } from "vitest";
import { lightFetch } from "@/lib/light-auth/api-fetch";
import { skillCatalogApi } from "../skillCatalogApi";

vi.mock("@/lib/light-auth/api-fetch", () => ({ lightFetch: vi.fn() }));
vi.mock("@/stores/auth", () => ({ readCurrentOrg: () => ({ slug: "acme" }) }));

describe("skillCatalogApi", () => {
  beforeEach(() => vi.clearAllMocks());

  it("projects tags from the REST catalog", async () => {
    vi.mocked(lightFetch).mockResolvedValue({
      skills: [{ slug: "video-editing", tags: ["editing", "video"] }],
      total: 1,
    });

    const result = await skillCatalogApi.list();

    expect(result.skills[0].tags).toEqual(["editing", "video"]);
  });

  it("sends tag-only PATCH updates", async () => {
    vi.mocked(lightFetch).mockResolvedValue({
      skill: { slug: "video-editing", tags: ["curated", "video"] },
    });

    const result = await skillCatalogApi.update("video-editing", {
      tags: ["curated", "video"],
    });

    expect(lightFetch).toHaveBeenCalledWith(
      "/api/v1/orgs/acme/authored-skills/video-editing",
      {
        method: "PATCH",
        body: { tags: ["curated", "video"] },
        authenticated: true,
      },
    );
    expect(result.tags).toEqual(["curated", "video"]);
  });
});
