import { beforeEach, describe, expect, it, vi } from "vitest";
import { authenticatedFetch } from "./identity";
import { listModelResources } from "./modelConfigsApi";

vi.mock("./identity", () => ({
  authenticatedFetch: vi.fn(),
}));

const authenticatedFetchMock = vi.mocked(authenticatedFetch);

describe("model resource API", () => {
  beforeEach(() => {
    authenticatedFetchMock.mockReset();
  });

  it("returns selectable model resources", async () => {
    authenticatedFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          object: "list",
          data: [
            {
              id: 42,
              name: "Codex",
              provider_key: "openai",
              model: "gpt-5.5",
              is_default: true,
            },
          ],
        }),
        { status: 200 },
      ),
    );

    await expect(listModelResources()).resolves.toEqual([
      {
        id: 42,
        name: "Codex",
        provider_key: "openai",
        model: "gpt-5.5",
        is_default: true,
      },
    ]);
  });

  it("does not convert request failures into an empty resource list", async () => {
    authenticatedFetchMock.mockResolvedValue(new Response("database unavailable", { status: 503 }));

    await expect(listModelResources()).rejects.toThrow("database unavailable");
  });
});
