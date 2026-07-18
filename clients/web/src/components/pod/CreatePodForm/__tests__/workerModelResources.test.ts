import { describe, expect, it } from "vitest";
import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";
import {
  agentRequiresModelResource,
  compatibleModelResources,
  compatibleToolModelResources,
} from "../workerModelResources";

const geminiProvider: ProviderDefinition = {
  key: "gemini",
  displayName: "Gemini",
  modalities: ["chat"],
  credentialFields: [],
  defaultBaseUrl: "https://generativelanguage.googleapis.com",
  protocolAdapter: "gemini",
  supportsCustomEndpoint: false,
  supportsModelDiscovery: false,
};

const geminiResource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 1,
    ownerScope: "user",
    identifier: "gemini-main",
    providerKey: "gemini",
    name: "Gemini",
    baseUrl: "https://generativelanguage.googleapis.com",
    configuredFields: ["api_key"],
    status: "valid",
    isEnabled: true,
    validationError: "",
    canManage: true,
    resources: [],
  },
  resource: {
    id: 42,
    providerConnectionId: 1,
    identifier: "gemini-pro",
    modelId: "gemini-pro",
    displayName: "Gemini Pro",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: ["chat"],
    status: "valid",
    isEnabled: true,
    validationError: "",
  },
};

const minimaxProvider: ProviderDefinition = {
  ...geminiProvider,
  key: "minimax",
  displayName: "MiniMax",
  protocolAdapter: "minimax",
};

describe("workerModelResources", () => {
  it("allows selectable Gemini resources when exact model injection is supported", () => {
    expect(agentRequiresModelResource("gemini-cli")).toBe(true);
    expect(compatibleModelResources("gemini-cli", [geminiResource], [geminiProvider])).toEqual([
      geminiResource,
    ]);
  });

  it("allows OpenAI-compatible resources for video-studio", () => {
    const openAIProvider: ProviderDefinition = {
      ...geminiProvider,
      key: "openai",
      protocolAdapter: "openai-compatible",
    };
    const openAIResource: EffectiveResource = {
      ...geminiResource,
      connection: {
        ...geminiResource.connection!,
        providerKey: "openai",
      },
    };

    expect(agentRequiresModelResource("video-studio")).toBe(true);
    expect(
      compatibleModelResources("video-studio", [openAIResource], [openAIProvider]),
    ).toEqual([openAIResource]);
  });

  it.each(["openclaw", "hermes"])("%s only accepts its declared OpenAI-compatible resource", (agentSlug) => {
    const providers: ProviderDefinition[] = [
      { ...geminiProvider, key: "openai", protocolAdapter: "openai-compatible" },
      { ...geminiProvider, key: "anthropic", protocolAdapter: "anthropic" },
      geminiProvider,
    ];
    const resources: EffectiveResource[] = providers.map((provider, index) => ({
      ...geminiResource,
      connection: {
        ...geminiResource.connection!,
        id: index + 1,
        providerKey: provider.key,
      },
      resource: {
        ...geminiResource.resource!,
        id: index + 10,
        providerConnectionId: index + 1,
      },
    }));

    expect(agentRequiresModelResource(agentSlug)).toBe(true);
    expect(compatibleModelResources(agentSlug, resources, providers)).toEqual([resources[0]]);
  });

  it("allows selectable MiniMax resources for MiniMax CLI", () => {
    const minimaxResource: EffectiveResource = {
      ...geminiResource,
      connection: {
        ...geminiResource.connection!,
        providerKey: "minimax",
      },
    };

    expect(agentRequiresModelResource("minimax-cli")).toBe(true);
    expect(compatibleModelResources("minimax-cli", [minimaxResource], [minimaxProvider])).toEqual([
      minimaxResource,
    ]);
  });

  it("does not offer MiniMax chat resources to Seedance", () => {
    const minimaxResource: EffectiveResource = {
      ...geminiResource,
      connection: {
        ...geminiResource.connection!,
        providerKey: "minimax",
      },
    };

    expect(
      compatibleModelResources("seedance-expert", [minimaxResource], [minimaxProvider]),
    ).toEqual([]);
  });

  it("allows declared Doubao and Sub2API Seedance video resources", () => {
    const video = {
      ...geminiResource,
      connection: {
        ...geminiResource.connection!,
        providerKey: "doubao",
      },
      resource: {
        ...geminiResource.resource!,
        id: 77,
        modelId: "doubao-seedance-2-0-260128",
        modalities: ["video"],
        capabilities: ["video-generation"],
      },
    };
    const languageModelMarkedAsVideo = {
      ...video,
      resource: {
        ...video.resource!,
        id: 78,
        modelId: "doubao-seed-1-8-251228",
      },
    };
    const sub2apiVideo = {
      ...video,
      connection: {
        ...video.connection!,
        id: 2,
        providerKey: "sub2api-seedance",
        name: "Sub2API Seedance",
      },
      resource: {
        ...video.resource!,
        id: 79,
        providerConnectionId: 2,
        modelId: "doubao-seedance-2-0-260128",
      },
    };
    const sub2apiInvalidModelID = {
      ...sub2apiVideo,
      resource: {
        ...sub2apiVideo.resource!,
        id: 80,
        modelId: "doubao-seedance-2-0-260128-preview",
      },
    };

    expect(compatibleToolModelResources({
      role: "seedance-video",
      provider_keys: ["doubao", "sub2api-seedance"],
      protocol_adapters: ["openai-compatible", "ark-seedance"],
      modality: "video",
      capability: "video-generation",
    }, [
      geminiResource,
      languageModelMarkedAsVideo,
      video,
      sub2apiVideo,
      sub2apiInvalidModelID,
    ])).toEqual([
      video,
      sub2apiVideo,
    ]);
  });
});
