import { beforeEach, describe, expect, it, vi } from "vitest";
import { lightFetch } from "@/lib/light-auth/api-fetch";
import { quickTaskApi } from "../quickTaskApi";

vi.mock("@/lib/light-auth/api-fetch", () => ({
  lightFetch: vi.fn(),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "team-alpha" }),
}));

describe("quickTaskApi", () => {
  beforeEach(() => {
    vi.mocked(lightFetch).mockReset();
  });

  it("submits only the Worker plan ID", async () => {
    vi.mocked(lightFetch).mockResolvedValue({
      pod_key: "7-standalone-12345678",
      status: "queued",
    });

    await quickTaskApi.create({
      plan_id: "11111111-1111-4111-8111-111111111111",
    });

    expect(lightFetch).toHaveBeenCalledWith(
      "/api/v1/orgs/team-alpha/quick-tasks",
      {
        method: "POST",
        body: {
          plan_id: "11111111-1111-4111-8111-111111111111",
        },
        authenticated: true,
      },
    );
  });
});
