import { vi } from "vitest";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import type { WorkerCreateController } from "../../hooks/workerCreateController";
import type { WorkerCreateDraftState } from "../../hooks/workerCreateDraft";

export const mockPatchDraft = vi.fn();
export const mockChangeWorkerType = vi.fn();
export const mockSetLifecycle = vi.fn();
export const mockSetFillPrompt = vi.fn();
export const mockFillWithAI = vi.fn(async () => undefined);
export const mockGoToStep = vi.fn(async () => undefined);
export const mockRunPreflight = vi.fn(async () => null);
export const mockCreateWorker = vi.fn(async () => null);
export const mockReset = vi.fn();

export function controllerFixture(overrides: {
  state?: Partial<WorkerCreateDraftState>;
  controller?: Partial<WorkerCreateController>;
} = {}): WorkerCreateController {
  const state: WorkerCreateDraftState = {
    instanceId: "worker-create-test",
    step: 1,
    fillPrompt: "",
    draft: completeDraft(),
    fill: { status: "idle" },
    fillRequestId: null,
    preflight: { status: "idle" },
    preflightRequestId: null,
    create: { status: "idle" },
    ...overrides.state,
  };
  return {
    state,
    options: { status: "ready", data: createOptions() },
    modelResources: { status: "ready", data: [modelResource()] },
    toolModelResources: { status: "ready", data: [] },
    runtimeBundles: { status: "ready", data: [] },
    credentialBundles: { status: "ready", data: [] },
    configBundles: { status: "ready", data: [] },
    skills: { status: "ready", data: [] },
    repositories: [mockRepository],
    validity: {
      runtime: true,
      typeConfig: true,
      workspace: true,
      accessible: () => true,
    },
    patchDraft: mockPatchDraft,
    changeWorkerType: mockChangeWorkerType,
    setLifecycle: mockSetLifecycle,
    setFillPrompt: mockSetFillPrompt,
    fillWithAI: mockFillWithAI,
    goToStep: mockGoToStep,
    runPreflight: mockRunPreflight,
    createWorker: mockCreateWorker,
    reset: mockReset,
    ...overrides.controller,
  };
}

export function completeDraft(): WorkerSpecDraft {
  return {
    model_resource_id: 42,
    tool_model_resource_ids: {},
    worker_type_slug: "codex-cli",
    runtime_image_id: 11,
    placement_policy: "automatic",
    compute_target_id: 21,
    deployment_mode: "pooled",
    resource_profile_id: 31,
    type_schema_version: 1,
    type_config_values: {},
    secret_refs: [],
    interaction_mode: "acp",
    automation_level: "autonomous",
    repository_id: 51,
    branch: "main",
    skill_ids: [],
    knowledge_mounts: [],
    env_bundle_ids: [],
    config_bundle_ids: [],
    instructions: "",
    initial_task: "Fix the failing test.",
    termination_policy: "manual",
    idle_timeout_minutes: 0,
    alias: "",
    options_revision: "runtime-catalog-1",
  };
}

export function createOptions(): WorkerCreateOptions {
  return {
    revision: "runtime-catalog-1",
    worker_types: [{
      slug: "codex-cli",
      name: "Codex CLI",
      description: "",
      schema_version: 1,
      config_schema: { version: 1, fields: {} },
      requires_model_resource: true,
      tool_model_requirements: [],
      selectable: true,
      blocking_reason: "",
    }],
    runtime_images: [{
      id: 11,
      slug: "codex-stable",
      name: "Codex stable",
      reference: "registry/codex@sha256:test",
      digest: "sha256:test",
      worker_type_slugs: ["codex-cli"],
      selectable: true,
      blocking_reason: "",
    }],
    compute_targets: [{
      id: 21,
      slug: "runner-pool",
      name: "Runner pool",
      kind: "runner-pool",
      supports_pooled: true,
      supports_dedicated: false,
      selectable: true,
      blocking_reason: "",
    }],
    deployment_modes: [{
      value: "pooled",
      name: "Pooled",
      selectable: true,
      blocking_reason: "",
    }],
    resource_profiles: [{
      id: 31,
      slug: "standard",
      name: "Standard",
      cpu_request_millicpu: 200,
      cpu_limit_millicpu: 1000,
      memory_request_bytes: 268435456,
      memory_limit_bytes: 1073741824,
      storage_request_bytes: 10737418240,
      storage_limit_bytes: 10737418240,
      selectable: true,
      blocking_reason: "",
    }],
  };
}

export const mockRepository = {
  id: 51,
  organization_id: 1,
  provider_type: "github",
  provider_base_url: "https://github.com",
  http_clone_url: "https://github.com/org/repo.git",
  external_id: "org-repo",
  name: "repo",
  slug: "org-repo",
  default_branch: "main",
  visibility: "organization",
  is_active: true,
  created_at: "2026-07-10T00:00:00Z",
  updated_at: "2026-07-10T00:00:00Z",
};

export function modelResource(): EffectiveResource {
  return {
    selectable: true,
    blockingReason: "",
    connection: {
      id: 1,
      ownerScope: "organization",
      identifier: "openai",
      providerKey: "openai",
      name: "OpenAI",
      baseUrl: "https://api.openai.com/v1",
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
      identifier: "gpt-5",
      modelId: "gpt-5",
      displayName: "GPT-5",
      modalities: ["chat"],
      capabilities: ["text-generation"],
      defaultModalities: ["chat"],
      status: "valid",
      isEnabled: true,
      validationError: "",
    },
  };
}
