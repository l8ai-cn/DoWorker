import { beforeEach, describe, expect, it, vi } from "vitest";

import { lightFetch } from "@/lib/light-auth/api-fetch";
import { listExpertMarketSubmissions } from "./expertMarketApi";

vi.mock("@/lib/light-auth/api-fetch", () => ({
  lightFetch: vi.fn(),
}));
vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "studio" }),
}));

describe("listExpertMarketSubmissions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("loads every page reported by total", async () => {
    const firstPage = Array.from({ length: 100 }, (_, index) => release(index + 1));
    const secondPage = Array.from({ length: 25 }, (_, index) => release(index + 101));
    vi.mocked(lightFetch)
      .mockResolvedValueOnce({ releases: firstPage, total: 125 })
      .mockResolvedValueOnce({ releases: secondPage, total: 125 });

    const result = await listExpertMarketSubmissions();

    expect(result.releases).toHaveLength(125);
    expect(result.total).toBe(125);
    expect(lightFetch).toHaveBeenNthCalledWith(
      1,
      "/api/v1/orgs/studio/marketplace/submissions",
      { authenticated: true, query: { limit: 100, offset: 0 } },
    );
    expect(lightFetch).toHaveBeenNthCalledWith(
      2,
      "/api/v1/orgs/studio/marketplace/submissions",
      { authenticated: true, query: { limit: 100, offset: 100 } },
    );
  });

  it("returns the application slug supplied by the submissions API", async () => {
    vi.mocked(lightFetch).mockResolvedValueOnce({
      releases: [release(1)],
      total: 1,
    });

    const result = await listExpertMarketSubmissions();

    expect(result.releases[0].application_slug).toBe("video-production");
    expect(lightFetch).toHaveBeenCalledTimes(1);
  });
});

function release(id: number) {
  return {
    id,
    application_id: 4,
    application_slug: "video-production",
    source_expert_id: 7,
    version: id,
    status: "published" as const,
    name: "Video Director",
    summary: "Plans video production",
    description: "",
    category: "video",
    icon: "film",
    tags: [],
    outcomes: [],
    created_at: "2026-07-15T08:00:00Z",
  };
}
