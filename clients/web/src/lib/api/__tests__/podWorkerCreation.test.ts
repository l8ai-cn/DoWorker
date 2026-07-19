import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  FillWorkerDraftRequestSchema,
  FillWorkerDraftResponseSchema,
  ListWorkerCreateOptionsRequestSchema,
  ListWorkerCreateOptionsResponseSchema,
  PreflightWorkerRequestSchema,
  PreflightWorkerResponseSchema,
  WorkerSpecDraftSchema,
} from "@proto/pod/v1/worker_creation_pb";
import {
  CreatePodRequestSchema,
  CreatePodResponseSchema,
  PodSchema,
} from "@proto/pod/v1/pod_pb";

vi.mock("@/lib/wasm-core", () => ({
  getPodService: vi.fn(),
}));

import { getPodService } from "@/lib/wasm-core";
import { createPod } from "../connect/podConnect";
import {
  fillWorkerDraft,
  listWorkerCreateOptions,
  preflightWorker,
  workerDraftFromProto,
  type WorkerSpecDraft,
} from "../connect/podWorkerCreationConnect";

const service = {
  create_pod_connect: vi.fn(),
  list_worker_create_options_connect: vi.fn(),
  preflight_worker_connect: vi.fn(),
  fill_worker_draft_connect: vi.fn(),
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getPodService).mockReturnValue(
    service as unknown as ReturnType<typeof getPodService>,
  );
});

describe("worker creation Connect boundary", () => {
  it("round-trips options, full drafts, preflight, fill, and CreatePod", async () => {
    service.list_worker_create_options_connect.mockResolvedValue(
      toBinary(
        ListWorkerCreateOptionsResponseSchema,
        create(ListWorkerCreateOptionsResponseSchema, {
          revision: "rev-7",
          workerTypes: [{
            slug: "codex",
            name: "Codex",
            description: "Coding worker",
            schemaVersion: 3,
            configSchemaJson: '{"version":3,"fields":[]}',
            requiresModelResource: true,
            modelProtocolAdapters: ["openai-compatible", "anthropic"],
            toolModelRequirements: [{
              role: "seedance-video",
              providerKeys: ["doubao"],
              protocolAdapters: ["openai-compatible"],
              modality: "video",
              capability: "video-generation",
            }],
            credentialRequirements: [{
              id: "api-key",
              sourceKind: "environment_bundle",
              sourceRef: "OPENAI_API_KEY",
              targetKind: "environment",
              targetName: "OPENAI_API_KEY",
            }],
            configDocumentRequirements: [{
              documentId: "settings",
              format: "json",
              targetPath: "DO_AGENT_SETTINGS",
              required: true,
            }],
            selectable: true,
          }],
          runtimeImages: [{
            id: BigInt(12),
            slug: "codex-stable",
            name: "Codex stable",
            reference: "registry/codex@sha256:abc",
            digest: "sha256:abc",
            workerTypeSlugs: ["codex"],
            selectable: true,
          }],
          computeTargets: [{
            id: BigInt(21),
            slug: "local",
            name: "Local",
            kind: "runner",
            supportsPooled: true,
            supportsDedicated: false,
            selectable: true,
          }],
          deploymentModes: [{
            value: "pooled",
            name: "Pooled",
            selectable: true,
          }],
          resourceProfiles: [{
            id: BigInt(31),
            slug: "standard",
            name: "Standard",
            cpuRequestMillicpu: 500,
            cpuLimitMillicpu: 1000,
            memoryRequestBytes: BigInt(536870912),
            memoryLimitBytes: BigInt(1073741824),
            selectable: true,
          }],
        }),
      ),
    );
    const options = await listWorkerCreateOptions("acme", {
      worker_type_slug: "codex",
      compute_target_id: 21,
      deployment_mode: "pooled",
    });
    expect(options).toMatchObject({
      revision: "rev-7",
      worker_types: [{
        slug: "codex",
        config_schema: { version: 3, fields: [] },
        requires_model_resource: true,
        model_protocol_adapters: ["openai-compatible", "anthropic"],
        tool_model_requirements: [{
          role: "seedance-video",
          provider_keys: ["doubao"],
          protocol_adapters: ["openai-compatible"],
        }],
        credential_requirements: [{
          id: "api-key",
          source_kind: "environment_bundle",
          source_ref: "OPENAI_API_KEY",
          target_kind: "environment",
          target_name: "OPENAI_API_KEY",
        }],
        config_document_requirements: [{
          document_id: "settings",
          format: "json",
          target_path: "DO_AGENT_SETTINGS",
          required: true,
        }],
      }],
      runtime_images: [{ id: 12, worker_type_slugs: ["codex"] }],
      compute_targets: [{ id: 21, supports_pooled: true }],
      resource_profiles: [{ id: 31, memory_limit_bytes: 1073741824 }],
    });
    const optionsRequest = fromBinary(
      ListWorkerCreateOptionsRequestSchema,
      service.list_worker_create_options_connect.mock.calls[0][0],
    );
    expect(optionsRequest).toMatchObject({
      orgSlug: "acme",
      workerTypeSlug: "codex",
      computeTargetId: BigInt(21),
      deploymentMode: "pooled",
    });

    const draft = fullDraft();
    service.preflight_worker_connect.mockResolvedValue(
      toBinary(
        PreflightWorkerResponseSchema,
        create(PreflightWorkerResponseSchema, {
          issues: [{
            code: "repository.branch.missing",
            field: "branch",
            message: "Branch is required",
            severity: "error",
          }],
          resolvedSpecJson: '{"version":1}',
          optionsRevision: "rev-7",
        }),
      ),
    );
    const preflight = await preflightWorker("acme", draft);
    expect(preflight).toEqual({
      issues: [{
        code: "repository.branch.missing",
        field: "branch",
        message: "Branch is required",
        severity: "error",
      }],
      resolved_spec_json: '{"version":1}',
      options_revision: "rev-7",
    });
    const preflightRequest = fromBinary(
      PreflightWorkerRequestSchema,
      service.preflight_worker_connect.mock.calls[0][0],
    );
    expect(preflightRequest.draft).toMatchObject({
      modelResourceId: BigInt(9),
      toolModelResourceIds: {
        "seedance-video": BigInt(10),
        "video-generator": BigInt(91),
      },
      runtimeImageId: BigInt(12),
      repositoryId: BigInt(44),
      skillIds: [BigInt(51), BigInt(52)],
      envBundleIds: [BigInt(71)],
      configDocumentBindings: [{
        documentId: "settings",
        configBundleId: BigInt(72),
      }],
      sourceExpertId: BigInt(81),
      typeConfigValuesJson: '{"temperature":0.2,"nested":{"enabled":true}}',
    });

    service.fill_worker_draft_connect.mockResolvedValue(
      toBinary(
        FillWorkerDraftResponseSchema,
        create(FillWorkerDraftResponseSchema, {
          draft: create(WorkerSpecDraftSchema, {
            ...preflightRequest.draft!,
            alias: "AI filled",
          }),
          issues: [],
        }),
      ),
    );
    const filled = await fillWorkerDraft("acme", "Create a coding worker", draft);
    expect(filled.draft).toEqual({ ...draft, alias: "AI filled" });
    const fillRequest = fromBinary(
      FillWorkerDraftRequestSchema,
      service.fill_worker_draft_connect.mock.calls[0][0],
    );
    expect(fillRequest.generationModelResourceId).toBe(BigInt(0));

    service.create_pod_connect.mockResolvedValue(
      toBinary(
        CreatePodResponseSchema,
        create(CreatePodResponseSchema, {
          pod: create(PodSchema, {
            id: BigInt(1),
            podKey: "worker-1",
            status: "initializing",
            agentStatus: "idle",
            agentSlug: "codex",
            createdAt: "2026-07-10T00:00:00Z",
            updatedAt: "2026-07-10T00:00:00Z",
          }),
        }),
      ),
    );
    await createPod("acme", { worker_spec: draft });
    const createRequest = fromBinary(
      CreatePodRequestSchema,
      service.create_pod_connect.mock.calls[0][0],
    );
    expect(createRequest.workerSpec).toEqual(preflightRequest.draft);
    expect(createRequest.agentSlug).toBe("");
  });

  it("rejects unsafe identifiers and missing filled drafts", async () => {
    expect(() =>
      workerDraftFromProto(create(WorkerSpecDraftSchema, {
        modelResourceId: BigInt(Number.MAX_SAFE_INTEGER) + BigInt(1),
      })),
    ).toThrow("unsafe model_resource_id");

    service.fill_worker_draft_connect.mockResolvedValue(
      toBinary(
        FillWorkerDraftResponseSchema,
        create(FillWorkerDraftResponseSchema, {}),
      ),
    );
    await expect(fillWorkerDraft("acme", "Create worker")).rejects.toThrow(
      "worker draft response is missing draft",
    );
  });
});

function fullDraft(): WorkerSpecDraft {
  return {
    model_resource_id: 9,
    tool_model_resource_ids: {
      "seedance-video": 10,
      "video-generator": 91,
    },
    worker_type_slug: "codex",
    runtime_image_id: 12,
    placement_policy: "preferred",
    compute_target_id: 21,
    deployment_mode: "pooled",
    resource_profile_id: 31,
    type_schema_version: 3,
    type_config_values: { temperature: 0.2, nested: { enabled: true } },
    secret_refs: [{ field: "api_key", kind: "secret", id: 41 }],
    interaction_mode: "terminal",
    automation_level: "autonomous",
    repository_id: 44,
    branch: "main",
    skill_ids: [51, 52],
    knowledge_mounts: [{ knowledge_base_id: 61, mode: "read_only" }],
    env_bundle_ids: [71],
    config_document_bindings: [{
      document_id: "settings",
      config_bundle_id: 72,
    }],
    instructions: "Follow the repository rules.",
    initial_task: "Implement the requested change.",
    termination_policy: "idle",
    idle_timeout_minutes: 30,
    alias: "Coding worker",
    source_expert_id: 81,
    options_revision: "rev-7",
  };
}
