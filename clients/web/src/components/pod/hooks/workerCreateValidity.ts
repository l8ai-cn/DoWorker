import type { WorkerCreateOptions, WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { AsyncState, WorkerCreateStepId } from "./workerCreateDraft";

export interface WorkerCreateValidity {
  runtime: boolean;
  typeConfig: boolean;
  workspace: boolean;
  accessible: (step: WorkerCreateStepId) => boolean;
}

export function workerCreateValidity(
  draft: WorkerSpecDraft,
  options: AsyncState<WorkerCreateOptions>,
  dependenciesReady: boolean,
): WorkerCreateValidity {
  const runtime = options.status === "ready" && runtimeSelectionsValid(draft, options.data);
  const typeConfig = runtime && draft.type_schema_version > 0;
  const workspace = typeConfig && dependenciesReady && workspaceValid(draft);
  return {
    runtime,
    typeConfig,
    workspace,
    accessible: (step) => {
      if (step === 1) return true;
      if (step === 2) return runtime;
      if (step === 3) return typeConfig;
      return workspace;
    },
  };
}

function runtimeSelectionsValid(
  draft: WorkerSpecDraft,
  options: WorkerCreateOptions,
): boolean {
  const workerType = options.worker_types.find(
    (option) => option.slug === draft.worker_type_slug,
  );
  return Boolean(
    workerType?.selectable &&
      (!workerType.requires_model_resource || draft.model_resource_id > 0) &&
      workerType.tool_model_requirements.every(
        (requirement) => (draft.tool_model_resource_ids[requirement.role] ?? 0) > 0,
      ) &&
      selectable(options.runtime_images, draft.runtime_image_id, (item) => item.id) &&
      selectable(options.compute_targets, draft.compute_target_id, (item) => item.id) &&
      selectable(options.deployment_modes, draft.deployment_mode, (item) => item.value) &&
      (draft.custom_resources
        ? customResourcesValid(draft.custom_resources)
        : selectable(options.resource_profiles, draft.resource_profile_id, (item) => item.id)),
  );
}

function customResourcesValid(resources: NonNullable<WorkerSpecDraft["custom_resources"]>): boolean {
  return resources.cpu_request_millicpu > 0 &&
    resources.cpu_limit_millicpu >= resources.cpu_request_millicpu &&
    resources.memory_request_bytes > 0 &&
    resources.memory_limit_bytes >= resources.memory_request_bytes &&
    resources.storage_request_bytes > 0 &&
    resources.storage_limit_bytes >= resources.storage_request_bytes;
}

function workspaceValid(draft: WorkerSpecDraft): boolean {
  if (draft.repository_id && !draft.branch.trim()) return false;
  if (draft.termination_policy === "idle" && draft.idle_timeout_minutes <= 0) {
    return false;
  }
  return true;
}

function selectable<T extends { selectable: boolean }, V>(
  options: T[],
  value: V,
  pick: (item: T) => V,
): boolean {
  return Boolean(options.find((item) => pick(item) === value)?.selectable);
}
