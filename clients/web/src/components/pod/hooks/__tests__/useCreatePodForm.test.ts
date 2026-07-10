import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import * as podConnect from "@/lib/api/facade/podConnect";
import * as envBundleConnect from "@/lib/api/facade/envBundleConnect";
import * as aiResourceConnect from "@/lib/api/facade/aiResourceConnect";
import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";

vi.mock("@/lib/api/facade/envBundleConnect");
const mockListEnvBundles = vi.mocked(envBundleConnect.listEnvBundles);

vi.mock("@/lib/api/facade/aiResourceConnect");
const mockGetCatalog = vi.mocked(aiResourceConnect.getCatalog);
const mockListPersonalEffectiveResources = vi.mocked(aiResourceConnect.listPersonalEffectiveResources);
const mockListOrganizationEffectiveResources = vi.mocked(aiResourceConnect.listOrganizationEffectiveResources);

vi.mock("@/lib/api/facade/podConnect");
const mockCreatePod = vi.mocked(podConnect.createPod);

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "test-org" }),
  useAuthStore: () => ({}),
}));

vi.mock("@/stores/podCreation", () => ({
  usePodCreationStore: () => ({
    lastAgentSlug: null,
    lastRepositoryId: null,
    lastRuntimeBundleNames: [],
    lastBranchName: null,
    lastSkillSlugs: [],
    setLastChoices: vi.fn(),
    clearLastChoices: vi.fn(),
    _hasHydrated: true,
    setHasHydrated: vi.fn(),
  }),
}));

import { useCreatePodForm } from "../useCreatePodForm";

const legacyField = (prefix: string, suffix: string) => `${prefix}_${suffix}`;
import { useEnvBundles } from "../useCreatePodFormEffects";

const mockAgents = [
  { name: "Claude Code", slug: "claude-code", is_builtin: true, is_active: true },
];

const anthropicProvider: ProviderDefinition = {
  key: "anthropic",
  displayName: "Anthropic",
  modalities: ["chat"],
  credentialFields: [],
  defaultBaseUrl: "https://api.anthropic.com",
  protocolAdapter: "anthropic",
  supportsCustomEndpoint: false,
  supportsModelDiscovery: false,
};

const claudeResource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 1,
    ownerScope: "user",
    identifier: "anthropic-main",
    providerKey: "anthropic",
    name: "Anthropic",
    baseUrl: "https://api.anthropic.com",
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
    identifier: "claude-sonnet",
    modelId: "claude-sonnet",
    displayName: "Claude Sonnet",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: ["chat"],
    status: "valid",
    isEnabled: true,
    validationError: "",
  },
};

async function selectClaudeResource(result: { current: ReturnType<typeof useCreatePodForm> }) {
  act(() => {
    result.current.setSelectedAgent("claude-code");
  });
  await waitFor(() => expect(result.current.modelResources).toHaveLength(1));
  act(() => {
    result.current.setSelectedModelResourceId(42);
  });
}

describe("useCreatePodForm - bundle via agentfile_layer (SSOT)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListEnvBundles.mockResolvedValue({ items: [], total: 0 });
    mockGetCatalog.mockResolvedValue([anthropicProvider]);
    mockListPersonalEffectiveResources.mockResolvedValue([claudeResource]);
    mockListOrganizationEffectiveResources.mockResolvedValue([claudeResource]);
  });

  it("omits USE_ENV_BUNDLE from agentfile_layer when no bundle is selected", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);

    expect(result.current.selectedModelResourceId).toBe(42);
    expect(result.current.selectedRuntimeBundleNames).toEqual([]);

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).toHaveBeenCalledTimes(1);
    const [, createArg] = mockCreatePod.mock.calls[0];
    expect(createArg).toHaveProperty("model_resource_id", 42);
    expect(createArg).not.toHaveProperty(legacyField("credential", "profile_id"));
    expect(createArg).not.toHaveProperty(legacyField("virtual_api", "key_id"));
    expect(createArg).not.toHaveProperty(legacyField("model", "config_id"));
    const layer = createArg.agentfile_layer ?? "";
    expect(layer).not.toContain("USE_ENV_BUNDLE");
  });

  it("includes only explicit runtime USE_ENV_BUNDLE entries", async () => {
    const runtimeBundle = {
      $typeName: "proto.env_bundle.v1.EnvBundle" as const,
      id: BigInt(43), ownerScope: "user", ownerId: BigInt(1), agentSlug: "claude-code", name: "production-debug",
      kind: "runtime", kindPrimary: false, isActive: true,
      configuredFields: [], configuredValues: {},
      createdAt: "x", updatedAt: "x",
    };
    mockListEnvBundles.mockReset();
    mockListEnvBundles.mockResolvedValueOnce({ items: [runtimeBundle], total: 1 });
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);
    await act(async () => {});

    act(() => {
      result.current.setSelectedRuntimeBundleNames(["production-debug"]);
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).toHaveBeenCalledTimes(1);
    const [, createArg] = mockCreatePod.mock.calls[0];
    expect(createArg).not.toHaveProperty(legacyField("credential", "profile_id"));
    const layer: string = createArg.agentfile_layer ?? "";
    const useLines = layer.split("\n").filter((l) => l.startsWith("USE_ENV_BUNDLE"));
    expect(useLines).toEqual(['USE_ENV_BUNDLE "production-debug"']);
  });

  it("always sends agentfile_layer via API (SSOT)", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const [, createArg] = mockCreatePod.mock.calls[0];
    expect(createArg).toHaveProperty("agent_slug", "claude-code");
    expect(createArg).toHaveProperty("model_resource_id", 42);
    expect(createArg).not.toHaveProperty(legacyField("credential", "profile_id"));
    expect(createArg).not.toHaveProperty(legacyField("virtual_api", "key_id"));
    expect(createArg).not.toHaveProperty(legacyField("model", "config_id"));
    expect(createArg).not.toHaveProperty("repository_id");
    expect(createArg).not.toHaveProperty("interaction_mode");
    expect(createArg).not.toHaveProperty("branch_name");
    expect(createArg).not.toHaveProperty("prompt");
    expect(createArg).not.toHaveProperty("config_overrides");
  });

  it("emits ENV lines in agentfile_layer for custom env vars", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);
    act(() => {
      result.current.setCustomEnv([
        { id: "a", key: "FOO", value: "bar" },
        { id: "b", key: "HTTP_PROXY", value: "http://localhost:8080" },
      ]);
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const [, createArg] = mockCreatePod.mock.calls[0];
    const layer: string = createArg.agentfile_layer ?? "";
    expect(layer).toContain('ENV FOO = "bar"');
    expect(layer).toContain('ENV HTTP_PROXY = "http://localhost:8080"');
  });

  it("blocks submit when a custom env key is invalid", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);
    act(() => {
      result.current.setCustomEnv([{ id: "a", key: "bad-key", value: "x" }]);
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).not.toHaveBeenCalled();
    expect(result.current.validationErrors.env).toBeTruthy();
  });

  it("blocks submit when a model agent has no selected model resource", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    act(() => {
      result.current.setSelectedAgent("claude-code");
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).not.toHaveBeenCalled();
    expect(result.current.validationErrors.modelResource).toBeTruthy();
  });

  it("keeps the form invalid while runtime bundles are loading", async () => {
    let resolveBundles!: (value: { items: never[]; total: number }) => void;
    mockListEnvBundles.mockReturnValue(new Promise((resolve) => {
      resolveBundles = resolve;
    }));

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);
    expect(result.current.loadingBundles).toBe(true);
    expect(result.current.isValid).toBe(false);

    await act(async () => {
      resolveBundles({ items: [], total: 0 });
    });
  });

  it("keeps the form invalid when runtime bundles fail to load", async () => {
    mockListEnvBundles.mockRejectedValue(new Error("runtime bundle load failed"));

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    await selectClaudeResource(result);
    await waitFor(() => expect(result.current.bundleLoadError).toBe("runtime bundle load failed"));
    expect(result.current.isValid).toBe(false);
  });

  it("sends repository_id when a repository is selected", async () => {
    mockCreatePod.mockResolvedValue({
      pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } as never,
    });

    const repos = [{
      id: 7,
      organization_id: 1,
      provider_type: "github",
      provider_base_url: "https://github.com",
      http_clone_url: "https://github.com/org/repo.git",
      external_id: "org-repo",
      name: "repo",
      slug: "org/repo",
      default_branch: "main",
      visibility: "organization",
      is_active: true,
      created_at: "x",
      updated_at: "x",
    }];

    const { result } = renderHook(() => useCreatePodForm(mockAgents, repos));

    await selectClaudeResource(result);
    act(() => {
      result.current.setSelectedRepository(7);
    });
    await act(async () => {});
    act(() => {
      result.current.setSelectedBranch("develop");
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const [, createArg] = mockCreatePod.mock.calls[0];
    expect(createArg).toHaveProperty("repository_id", 7);
    expect(createArg.agentfile_layer).toContain('REPO "org/repo"');
    expect(createArg.agentfile_layer).toContain('BRANCH "develop"');
  });
});

describe("useEnvBundles", () => {
  it("ignores a stale response after the selected agent changes", async () => {
    type BundleListResponse = Awaited<ReturnType<typeof envBundleConnect.listEnvBundles>>;
    let resolveClaude!: (value: BundleListResponse) => void;
    let resolveCodex!: (value: BundleListResponse) => void;
    mockListEnvBundles.mockImplementation((args) => new Promise((resolve) => {
      const agentSlug = args?.agentSlug;
      if (agentSlug === "claude-code") {
        resolveClaude = resolve;
      } else {
        resolveCodex = resolve;
      }
    }) as ReturnType<typeof envBundleConnect.listEnvBundles>);

    const { result, rerender } = renderHook(
      ({ agentSlug }) => useEnvBundles(agentSlug),
      { initialProps: { agentSlug: "claude-code" as string | null } },
    );

    rerender({ agentSlug: "codex-cli" });
    await act(async () => {
      resolveCodex({
        items: [{
          $typeName: "proto.env_bundle.v1.EnvBundle",
          id: BigInt(2),
          ownerScope: "user",
          ownerId: BigInt(1),
          agentSlug: "codex-cli",
          name: "codex-runtime",
          kind: "runtime",
          kindPrimary: false,
          isActive: true,
          configuredFields: [],
          configuredValues: {},
          createdAt: "x",
          updatedAt: "x",
        }],
        total: 1,
      });
    });
    await waitFor(() => expect(result.current.envBundles[0]?.name).toBe("codex-runtime"));

    await act(async () => {
      resolveClaude({
        items: [{
          $typeName: "proto.env_bundle.v1.EnvBundle",
          id: BigInt(1),
          ownerScope: "user",
          ownerId: BigInt(1),
          agentSlug: "claude-code",
          name: "claude-runtime",
          kind: "runtime",
          kindPrimary: false,
          isActive: true,
          configuredFields: [],
          configuredValues: {},
          createdAt: "x",
          updatedAt: "x",
        }],
        total: 1,
      });
    });

    expect(result.current.envBundles[0]?.name).toBe("codex-runtime");
  });
});
