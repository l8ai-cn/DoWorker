import { create } from "@bufbuild/protobuf";
import {
  WorkerSpecDraftSchema,
  type WorkerSpecDraft as WorkerSpecDraftMessage,
} from "@proto/pod/v1/worker_creation_pb";

import { workerBigInt, workerNumber } from "./podWorkerCreationNumbers";
import type { WorkerSpecDraft } from "./podWorkerCreationTypes";

export function workerDraftToProto(draft: WorkerSpecDraft): WorkerSpecDraftMessage {
  return create(WorkerSpecDraftSchema, {
    modelResourceId: workerBigInt(draft.model_resource_id, "model_resource_id"),
    toolModelResourceIds: Object.fromEntries(
      Object.entries(draft.tool_model_resource_ids).map(([role, id]) => [
        role,
        workerBigInt(id, `tool_model_resource_ids.${role}`),
      ]),
    ),
    workerTypeSlug: draft.worker_type_slug,
    runtimeImageId: workerBigInt(draft.runtime_image_id, "runtime_image_id"),
    placementPolicy: draft.placement_policy,
    computeTargetId: workerBigInt(draft.compute_target_id, "compute_target_id"),
    deploymentMode: draft.deployment_mode,
    resourceProfileId: workerBigInt(draft.resource_profile_id, "resource_profile_id"),
    customResources: draft.custom_resources
      ? {
          cpuRequestMillicpu: draft.custom_resources.cpu_request_millicpu,
          cpuLimitMillicpu: draft.custom_resources.cpu_limit_millicpu,
          memoryRequestBytes: workerBigInt(
            draft.custom_resources.memory_request_bytes,
            "custom_resources.memory_request_bytes",
          ),
          memoryLimitBytes: workerBigInt(
            draft.custom_resources.memory_limit_bytes,
            "custom_resources.memory_limit_bytes",
          ),
          storageRequestBytes: workerBigInt(
            draft.custom_resources.storage_request_bytes,
            "custom_resources.storage_request_bytes",
          ),
          storageLimitBytes: workerBigInt(
            draft.custom_resources.storage_limit_bytes,
            "custom_resources.storage_limit_bytes",
          ),
        }
      : undefined,
    typeSchemaVersion: draft.type_schema_version,
    typeConfigValuesJson: JSON.stringify(draft.type_config_values),
    secretRefs: draft.secret_refs.map((reference) => ({
      field: reference.field,
      kind: reference.kind,
      id: workerBigInt(reference.id, `secret_refs.${reference.field}.id`),
    })),
    interactionMode: draft.interaction_mode,
    automationLevel: draft.automation_level,
    repositoryId:
      draft.repository_id === undefined
        ? undefined
        : workerBigInt(draft.repository_id, "repository_id"),
    branch: draft.branch,
    skillIds: draft.skill_ids.map((id) => workerBigInt(id, "skill_ids")),
    knowledgeMounts: draft.knowledge_mounts.map((mount) => ({
      knowledgeBaseId: workerBigInt(mount.knowledge_base_id, "knowledge_mounts.knowledge_base_id"),
      mode: mount.mode,
    })),
    envBundleIds: draft.env_bundle_ids.map((id) => workerBigInt(id, "env_bundle_ids")),
    configDocumentBindings: draft.config_document_bindings.map((binding) => ({
      documentId: binding.document_id,
      configBundleId: workerBigInt(
        binding.config_bundle_id,
        `config_document_bindings.${binding.document_id}.config_bundle_id`,
      ),
    })),
    instructions: draft.instructions,
    initialTask: draft.initial_task,
    terminationPolicy: draft.termination_policy,
    idleTimeoutMinutes: draft.idle_timeout_minutes,
    alias: draft.alias,
    sourceExpertId:
      draft.source_expert_id === undefined
        ? undefined
        : workerBigInt(draft.source_expert_id, "source_expert_id"),
    optionsRevision: draft.options_revision,
  });
}

export function workerDraftFromProto(draft: WorkerSpecDraftMessage): WorkerSpecDraft {
  return {
    model_resource_id: workerNumber(draft.modelResourceId, "model_resource_id"),
    tool_model_resource_ids: Object.fromEntries(
      Object.entries(draft.toolModelResourceIds).map(([role, id]) => [
        role,
        workerNumber(id, `tool_model_resource_ids.${role}`),
      ]),
    ),
    worker_type_slug: draft.workerTypeSlug,
    runtime_image_id: workerNumber(draft.runtimeImageId, "runtime_image_id"),
    placement_policy: draft.placementPolicy,
    compute_target_id: workerNumber(draft.computeTargetId, "compute_target_id"),
    deployment_mode: draft.deploymentMode,
    resource_profile_id: workerNumber(draft.resourceProfileId, "resource_profile_id"),
    custom_resources: draft.customResources
      ? {
          cpu_request_millicpu: draft.customResources.cpuRequestMillicpu,
          cpu_limit_millicpu: draft.customResources.cpuLimitMillicpu,
          memory_request_bytes: workerNumber(
            draft.customResources.memoryRequestBytes,
            "custom_resources.memory_request_bytes",
          ),
          memory_limit_bytes: workerNumber(
            draft.customResources.memoryLimitBytes,
            "custom_resources.memory_limit_bytes",
          ),
          storage_request_bytes: workerNumber(
            draft.customResources.storageRequestBytes,
            "custom_resources.storage_request_bytes",
          ),
          storage_limit_bytes: workerNumber(
            draft.customResources.storageLimitBytes,
            "custom_resources.storage_limit_bytes",
          ),
        }
      : undefined,
    type_schema_version: draft.typeSchemaVersion,
    type_config_values: parseTypeConfig(draft.typeConfigValuesJson),
    secret_refs: draft.secretRefs.map((reference) => ({
      field: reference.field,
      kind: reference.kind,
      id: workerNumber(reference.id, `secret_refs.${reference.field}.id`),
    })),
    interaction_mode: draft.interactionMode,
    automation_level: draft.automationLevel,
    repository_id:
      draft.repositoryId === undefined
        ? undefined
        : workerNumber(draft.repositoryId, "repository_id"),
    branch: draft.branch,
    skill_ids: draft.skillIds.map((id) => workerNumber(id, "skill_ids")),
    knowledge_mounts: draft.knowledgeMounts.map((mount) => ({
      knowledge_base_id: workerNumber(
        mount.knowledgeBaseId,
        "knowledge_mounts.knowledge_base_id",
      ),
      mode: mount.mode,
    })),
    env_bundle_ids: draft.envBundleIds.map((id) => workerNumber(id, "env_bundle_ids")),
    config_document_bindings: draft.configDocumentBindings.map((binding) => ({
      document_id: binding.documentId,
      config_bundle_id: workerNumber(
        binding.configBundleId,
        `config_document_bindings.${binding.documentId}.config_bundle_id`,
      ),
    })),
    instructions: draft.instructions,
    initial_task: draft.initialTask,
    termination_policy: draft.terminationPolicy,
    idle_timeout_minutes: draft.idleTimeoutMinutes,
    alias: draft.alias,
    source_expert_id:
      draft.sourceExpertId === undefined
        ? undefined
        : workerNumber(draft.sourceExpertId, "source_expert_id"),
    options_revision: draft.optionsRevision,
  };
}

function parseTypeConfig(raw: string): Record<string, unknown> {
  const value: unknown = JSON.parse(raw || "{}");
  if (value === null || Array.isArray(value) || typeof value !== "object") {
    throw new Error("worker type config must be a JSON object");
  }
  return value as Record<string, unknown>;
}
