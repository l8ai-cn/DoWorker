import { beforeEach, describe, expect, it, vi } from "vitest";

import { lightConnect } from "@/lib/light-auth/api-fetch";
import { listMarketplaceToolModelResources } from "./marketplace-tool-model-resources";

vi.mock("@/lib/light-auth/api-fetch", () => ({
  lightConnect: vi.fn(),
}));

const connectMock = vi.mocked(lightConnect);

describe("listMarketplaceToolModelResources", () => {
  beforeEach(() => connectMock.mockReset());

  it("uses canonical requirements to select Seedance video resources", async () => {
    connectMock.mockResolvedValue({
      resources: [
        {
          selectable: true,
          connection: {
            providerKey: "doubao",
            name: "Doubao",
            isEnabled: true,
          },
          resource: {
            id: "77",
            displayName: "Seedance 2.0",
            modelId: "doubao-seedance-2-0-260128",
            isEnabled: true,
            modalities: ["video"],
            capabilities: ["video-generation"],
          },
        },
        {
          selectable: true,
          connection: {
            providerKey: "doubao",
            name: "Doubao",
            isEnabled: true,
          },
          resource: {
            id: "78",
            displayName: "Other video",
            modelId: "doubao-video-other",
            isEnabled: true,
            modalities: ["video"],
            capabilities: ["video-generation"],
          },
        },
        {
          selectable: true,
          connection: {
            providerKey: "sub2api-seedance",
            name: "Sub2API Seedance",
            isEnabled: true,
          },
          resource: {
            id: "79",
            displayName: "Seedance 2.0",
            modelId: "doubao-seedance-2-0-260128",
            isEnabled: true,
            modalities: ["video"],
            capabilities: ["video-generation"],
          },
        },
        {
          selectable: true,
          connection: {
            providerKey: "sub2api-seedance",
            name: "Sub2API Seedance",
            isEnabled: true,
          },
          resource: {
            id: "80",
            displayName: "Legacy Sub2API model",
            modelId: "creative-video",
            isEnabled: true,
            modalities: ["video"],
            capabilities: ["video-generation"],
          },
        },
      ],
    });

    await expect(
      listMarketplaceToolModelResources("acme", "seedance-expert"),
    ).resolves.toEqual([
      {
        role: "seedance-video",
        resources: [
          { id: 77, label: "Doubao · Seedance 2.0" },
          { id: 79, label: "Sub2API Seedance · Seedance 2.0" },
        ],
      },
    ]);
    expect(connectMock).toHaveBeenCalledWith(
      expect.any(String),
      "ListOrganizationEffectiveResources",
      { orgSlug: "acme", modalities: ["video"] },
      { authenticated: true },
    );
  });

  it("does not call the API for workers without tool models", async () => {
    await expect(
      listMarketplaceToolModelResources("acme", "video-studio"),
    ).resolves.toEqual([]);
    expect(connectMock).not.toHaveBeenCalled();
  });
});
