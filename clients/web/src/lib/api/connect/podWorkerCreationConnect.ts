import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  FillWorkerDraftRequestSchema,
  FillWorkerDraftResponseSchema,
  PreflightWorkerRequestSchema,
  PreflightWorkerResponseSchema,
} from "@proto/pod/v1/worker_creation_pb";

import { getPodService } from "@/lib/wasm-core";
import { workerDraftFromProto, workerDraftToProto } from "./podWorkerDraftProto";
import type {
  WorkerDraftFillResult,
  WorkerPreflightIssue,
  WorkerPreflightResult,
  WorkerSpecDraft,
} from "./podWorkerCreationTypes";

export async function listWorkerCreateOptions(
  orgSlug: string,
  filter: WorkerCreateOptionsFilter = {},
): Promise<WorkerCreateOptions> {
  const request = create(ListWorkerCreateOptionsRequestSchema, {
    orgSlug,
    workerTypeSlug: filter.worker_type_slug,
    computeTargetId:
      filter.compute_target_id === undefined
        ? undefined
        : workerBigInt(filter.compute_target_id, "compute_target_id"),
    deploymentMode: filter.deployment_mode,
  });
  const responseBytes = await getPodService().list_worker_create_options_connect(
    toBinary(ListWorkerCreateOptionsRequestSchema, request),
  );
  const response = fromBinary(
    ListWorkerCreateOptionsResponseSchema,
    new Uint8Array(responseBytes),
  );
  return {
    revision: response.revision,
    worker_types: response.workerTypes.map((option) => ({
      slug: option.slug,
      name: option.name,
      description: option.description,
      schema_version: option.schemaVersion,
      config_schema: parseConfigSchema(option.configSchemaJson),
      supported_interaction_modes: option.supportedInteractionModes,
      requires_model_resource: option.requiresModelResource,
      model_protocol_adapters: option.modelProtocolAdapters,
      tool_model_requirements: option.toolModelRequirements.map((requirement) => ({
        role: requirement.role,
        provider_keys: requirement.providerKeys,
        protocol_adapters: requirement.protocolAdapters,
        modality: requirement.modality,
        capability: requirement.capability,
      })),
      credential_requirements: option.credentialRequirements.map((requirement) => ({
        id: requirement.id,
        source_kind: requirement.sourceKind,
        source_ref: requirement.sourceRef,
        target_kind: requirement.targetKind,
        target_name: requirement.targetName,
      })),
      config_document_requirements: option.configDocumentRequirements.map(
        (requirement) => ({
          document_id: requirement.documentId,
          format: requirement.format,
          target_path: requirement.targetPath,
          required: requirement.required,
        }),
      ),
      selectable: option.selectable,
      blocking_reason: option.blockingReason,
    })),
    runtime_images: response.runtimeImages.map((option) => ({
      id: workerNumber(option.id, "runtime_images.id"),
      slug: option.slug,
      name: option.name,
      reference: option.reference,
      digest: option.digest,
      worker_type_slugs: option.workerTypeSlugs,
      selectable: option.selectable,
      blocking_reason: option.blockingReason,
    })),
    compute_targets: response.computeTargets.map((option) => ({
      id: workerNumber(option.id, "compute_targets.id"),
      slug: option.slug,
      name: option.name,
      kind: option.kind,
      supports_pooled: option.supportsPooled,
      supports_dedicated: option.supportsDedicated,
      selectable: option.selectable,
      blocking_reason: option.blockingReason,
    })),
    deployment_modes: response.deploymentModes.map((option) => ({
      value: option.value,
      name: option.name,
      selectable: option.selectable,
      blocking_reason: option.blockingReason,
    })),
    resource_profiles: response.resourceProfiles.map((option) => ({
      id: workerNumber(option.id, "resource_profiles.id"),
      slug: option.slug,
      name: option.name,
      cpu_request_millicpu: option.cpuRequestMillicpu,
      cpu_limit_millicpu: option.cpuLimitMillicpu,
      memory_request_bytes: workerNumber(
        option.memoryRequestBytes,
        "resource_profiles.memory_request_bytes",
      ),
      memory_limit_bytes: workerNumber(
        option.memoryLimitBytes,
        "resource_profiles.memory_limit_bytes",
      ),
      storage_request_bytes: workerNumber(
        option.storageRequestBytes,
        "resource_profiles.storage_request_bytes",
      ),
      storage_limit_bytes: workerNumber(
        option.storageLimitBytes,
        "resource_profiles.storage_limit_bytes",
      ),
      gpu_request: option.gpuRequest,
      gpu_limit: option.gpuLimit,
      selectable: option.selectable,
      blocking_reason: option.blockingReason,
    })),
  };
}

export async function preflightWorker(
  orgSlug: string,
  draft: WorkerSpecDraft,
): Promise<WorkerPreflightResult> {
  const request = create(PreflightWorkerRequestSchema, {
    orgSlug,
    draft: workerDraftToProto(draft),
  });
  const responseBytes = await getPodService().preflight_worker_connect(
    toBinary(PreflightWorkerRequestSchema, request),
  );
  const response = fromBinary(
    PreflightWorkerResponseSchema,
    new Uint8Array(responseBytes),
  );
  return {
    issues: response.issues.map(preflightIssueFromProto),
    resolved_spec_json: response.resolvedSpecJson,
    options_revision: response.optionsRevision,
  };
}

export async function fillWorkerDraft(
  orgSlug: string,
  prompt: string,
  currentDraft?: WorkerSpecDraft,
): Promise<WorkerDraftFillResult> {
  const request = create(FillWorkerDraftRequestSchema, {
    orgSlug,
    prompt,
    currentDraft: currentDraft ? workerDraftToProto(currentDraft) : undefined,
  });
  const responseBytes = await getPodService().fill_worker_draft_connect(
    toBinary(FillWorkerDraftRequestSchema, request),
  );
  const response = fromBinary(
    FillWorkerDraftResponseSchema,
    new Uint8Array(responseBytes),
  );
  if (!response.draft) {
    throw new Error("worker draft response is missing draft");
  }
  return {
    draft: workerDraftFromProto(response.draft),
    issues: response.issues.map(preflightIssueFromProto),
  };
}

function preflightIssueFromProto(issue: {
  code: string;
  field: string;
  message: string;
  severity: string;
}): WorkerPreflightIssue {
  return {
    code: issue.code,
    field: issue.field,
    message: issue.message,
    severity: issue.severity,
  };
}

export { listWorkerCreateOptions } from "./podWorkerCreateOptionsConnect";
export { workerDraftFromProto, workerDraftToProto };
export type {
  WorkerCreateOptions,
  WorkerCreateOptionsFilter,
  WorkerConfigDocumentBinding,
  WorkerConfigDocumentRequirement,
  WorkerCredentialRequirement,
  WorkerDraftFillResult,
  WorkerPreflightIssue,
  WorkerPreflightResult,
  WorkerResourceRequest,
  WorkerSpecDraft,
  WorkerToolModelRequirement,
  WorkerTypeOption,
} from "./podWorkerCreationTypes";
