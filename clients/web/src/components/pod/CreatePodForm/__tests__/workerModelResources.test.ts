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

  it("uses Definition protocol adapters when they are provided", () => {
    const providers: ProviderDefinition[] = [
      { ...geminiProvider, key: "openai", protocolAdapter: "openai-compatible" },
      { ...geminiProvider, key: "anthropic", protocolAdapter: "anthropic" },
    ];
    const resources = providers.map((provider, index) => ({
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
    const compatibleForDefinition = compatibleModelResources as unknown as (
      agentSlug: string,
      values: EffectiveResource[],
      definitions: ProviderDefinition[],
      requirement: { required: boolean; protocolAdapters: string[] },
    ) => EffectiveResource[];

    expect(
      compatibleForDefinition(
        "new-definition-worker",
        resources,
        providers,
        { required: true, protocolAdapters: ["anthropic"] },
      ),
    ).toEqual([resources[1]]);
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

  it.each(["do-agent", "seedance-expert"])(
    "%s excludes MiniMax when its Definition only allows OpenAI-compatible and Anthropic models",
    (agentSlug) => {
      const providers: ProviderDefinition[] = [
        { ...geminiProvider, key: "openai", protocolAdapter: "openai-compatible" },
        { ...geminiProvider, key: "anthropic", protocolAdapter: "anthropic" },
        minimaxProvider,
      ];
      const resources = providers.map((provider, index) => ({
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

      expect(compatibleModelResources(agentSlug, resources, providers)).toEqual([
        resources[0],
        resources[1],
      ]);
    },
  );

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

  it("only allows the declared Doubao video-generation resource for Seedance", () => {
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

    expect(compatibleToolModelResources({
      role: "seedance-video",
      provider_keys: ["doubao"],
      protocol_adapters: ["openai-compatible"],
      modality: "video",
      capability: "video-generation",
    }, [geminiResource, languageModelMarkedAsVideo, video])).toEqual([video]);
  });
});
