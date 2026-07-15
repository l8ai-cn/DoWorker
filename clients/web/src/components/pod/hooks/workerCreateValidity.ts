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
  modelDependenciesReady = true,
): WorkerCreateValidity {
  const runtime = options.status === "ready" &&
    modelDependenciesReady &&
    runtimeSelectionsValid(draft, options.data);
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
    (item) => item.slug === draft.worker_type_slug && item.selectable,
  );
  if (!workerType) return false;
  const primaryModelValid = !workerType.requires_model_resource ||
    positiveInteger(draft.model_resource_id);
  const toolModelsValid = workerType.tool_model_requirements.every(
    (requirement) => positiveInteger(
      draft.tool_model_resource_ids[requirement.role],
    ),
  );
  return Boolean(
    primaryModelValid &&
      toolModelsValid &&
      selectable(options.runtime_images, draft.runtime_image_id, (item) => item.id) &&
      selectable(options.compute_targets, draft.compute_target_id, (item) => item.id) &&
      selectable(options.deployment_modes, draft.deployment_mode, (item) => item.value) &&
      selectable(options.resource_profiles, draft.resource_profile_id, (item) => item.id),
  );
}

function positiveInteger(value: number | undefined): boolean {
  return Number.isInteger(value) && Number(value) > 0;
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
