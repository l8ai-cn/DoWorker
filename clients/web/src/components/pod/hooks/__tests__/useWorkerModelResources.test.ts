import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import { useWorkerModelResources } from "../useWorkerModelResources";

const mockGetCatalog = vi.fn<() => Promise<ProviderDefinition[]>>();
const mockListOrganizationResources =
  vi.fn<() => Promise<EffectiveResource[]>>();

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "acme" }),
}));
vi.mock("@/lib/api/facade/aiResourceConnect", () => ({
  getCatalog: () => mockGetCatalog(),
  listOrganizationEffectiveResources: () => mockListOrganizationResources(),
  listPersonalEffectiveResources: () => mockListOrganizationResources(),
}));

describe("useWorkerModelResources", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCatalog.mockResolvedValue([
      provider("openai", "openai-compatible"),
      provider("anthropic", "anthropic"),
    ]);
  });

  it("does not expose resources loaded for the previous Worker type", async () => {
    const nextResources = deferred<EffectiveResource[]>();
    mockListOrganizationResources
      .mockResolvedValueOnce([resource(1, "openai")])
      .mockImplementationOnce(() => nextResources.promise);
    const { result, rerender } = renderHook(
      ({ workerType }) => useWorkerModelResources(workerType),
      { initialProps: { workerType: "codex-cli" } },
    );
    await waitFor(() =>
      expect(result.current.modelResources.map((item) => item.resource?.id))
        .toEqual([1]),
    );

    rerender({ workerType: "claude-code" });

    expect(result.current.loadingModelResources).toBe(true);
    expect(result.current.modelResources).toEqual([]);
    nextResources.resolve([resource(2, "anthropic")]);
    await waitFor(() =>
      expect(result.current.modelResources.map((item) => item.resource?.id))
        .toEqual([2]),
    );
  });
});

function provider(key: string, protocolAdapter: string): ProviderDefinition {
  return {
    key,
    displayName: key,
    modalities: ["chat"],
    credentialFields: [],
    defaultBaseUrl: "https://provider.example",
    protocolAdapter,
    supportsCustomEndpoint: false,
    supportsModelDiscovery: false,
  };
}

function resource(id: number, providerKey: string): EffectiveResource {
  return {
    selectable: true,
    blockingReason: "",
    connection: {
      id,
      ownerScope: "user",
      identifier: `${providerKey}-${id}`,
      providerKey,
      name: providerKey,
      baseUrl: "https://provider.example",
      configuredFields: ["api_key"],
      status: "valid",
      isEnabled: true,
      validationError: "",
      canManage: true,
      resources: [],
    },
    resource: {
      id,
      providerConnectionId: id,
      identifier: `model-${id}`,
      modelId: `model-${id}`,
      displayName: `Model ${id}`,
      modalities: ["chat"],
      capabilities: ["text-generation"],
      defaultModalities: ["chat"],
      status: "valid",
      isEnabled: true,
      validationError: "",
    },
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
