import { beforeEach, describe, expect, it, vi } from "vitest";

import { lightConnect } from "@/lib/light-auth/api-fetch";
import { listMarketplaceModelResources } from "./marketplace-model-resources";

vi.mock("@/lib/light-auth/api-fetch", () => ({
  lightConnect: vi.fn(),
}));

const connectMock = vi.mocked(lightConnect);

describe("listMarketplaceModelResources", () => {
  beforeEach(() => connectMock.mockReset());

  it("returns only selectable resources compatible with the expert agent", async () => {
    connectMock.mockImplementation(async (_service, method) => {
      if (method === "GetCatalog") {
        return {
          providers: [
            { key: "openai", protocolAdapter: "openai-compatible" },
            { key: "anthropic", protocolAdapter: "anthropic" },
          ],
        };
      }
      return {
        resources: [
          {
            selectable: true,
            connection: { providerKey: "openai", name: "OpenAI" },
            resource: {
              id: "42",
              displayName: "GPT 5.5",
              modelId: "gpt-5.5",
              isEnabled: true,
              modalities: ["chat"],
              capabilities: ["text-generation"],
            },
          },
          {
            selectable: true,
            connection: { providerKey: "anthropic", name: "Anthropic" },
            resource: {
              id: "43",
              displayName: "Claude",
              modelId: "claude-sonnet",
              isEnabled: true,
              modalities: ["chat"],
              capabilities: ["text-generation"],
            },
          },
        ],
      };
    });

    for (const agentSlug of ["codex-cli", "video-studio"]) {
      await expect(
        listMarketplaceModelResources("acme", agentSlug),
      ).resolves.toEqual([{ id: 42, label: "OpenAI · GPT 5.5" }]);
    }
  });

  it("rejects unsafe protobuf JSON identifiers", async () => {
    connectMock.mockResolvedValueOnce({
      providers: [{ key: "openai", protocolAdapter: "openai-compatible" }],
    });
    connectMock.mockResolvedValueOnce({
      resources: [
        {
          selectable: true,
          connection: { providerKey: "openai", name: "OpenAI" },
          resource: {
            id: "9007199254740993",
            displayName: "Unsafe",
            isEnabled: true,
            modalities: ["chat"],
            capabilities: ["text-generation"],
          },
        },
      ],
    });

    await expect(
      listMarketplaceModelResources("acme", "codex-cli"),
    ).rejects.toThrow("unsafe model resource id");
  });

  it("deduplicates concurrent requests for the same organization and agent", async () => {
    connectMock
      .mockResolvedValueOnce({
        providers: [{ key: "openai", protocolAdapter: "openai-compatible" }],
      })
      .mockResolvedValueOnce({ resources: [] });

    const first = listMarketplaceModelResources("acme", "video-studio");
    const second = listMarketplaceModelResources("acme", "video-studio");
    expect(connectMock).toHaveBeenCalledTimes(2);

    await expect(Promise.all([first, second])).resolves.toEqual([[], []]);
    expect(connectMock).toHaveBeenCalledTimes(2);
  });
});
