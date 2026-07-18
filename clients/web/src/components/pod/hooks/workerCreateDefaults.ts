import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
  WorkerTypeOption,
} from "@/lib/api/facade/podConnect";
import type { ProviderDefinition } from "@/lib/api/facade/aiResource";
import {
  compatibleWorkerModelResources,
  primaryModelRequirement,
  toolModelRequirement,
} from "../CreatePodForm/workerModelResourceCompatibility";

export function defaultWorkerDraftPatch(
  draft: WorkerSpecDraft,
  options: WorkerCreateOptions,
  preferredWorkerType?: string,
): Partial<WorkerSpecDraft> {
  const patch: Partial<WorkerSpecDraft> = {};
  const workerType = selectedOrFirst(
    options.worker_types,
    draft.worker_type_slug || preferredWorkerType || "",
    (option) => option.slug,
  );
  if (workerType) {
    if (!draft.worker_type_slug) {
      patch.worker_type_slug = workerType.slug;
    }
    if (draft.type_schema_version !== workerType.schema_version) {
      patch.type_schema_version = workerType.schema_version;
    }
  }
  const effectiveWorkerType = patch.worker_type_slug ?? draft.worker_type_slug;
  const image = selectedOrFirst(
    options.runtime_images.filter((option) =>
      option.worker_type_slugs.includes(effectiveWorkerType),
    ),
    String(draft.runtime_image_id),
    (option) => String(option.id),
  );
  if (!draft.runtime_image_id && image) patch.runtime_image_id = image.id;
  const target = selectedOrFirst(
    options.compute_targets,
    String(draft.compute_target_id),
    (option) => String(option.id),
  );
  if (!draft.compute_target_id && target) patch.compute_target_id = target.id;
  const deployment = selectedOrFirst(
    options.deployment_modes,
    draft.deployment_mode,
    (option) => option.value,
  );
  if (!draft.deployment_mode && deployment) {
    patch.deployment_mode = deployment.value;
  }
  const profile = selectedOrFirst(
    options.resource_profiles,
    String(draft.resource_profile_id),
    (option) => String(option.id),
  );
  if (!draft.resource_profile_id && profile) {
    patch.resource_profile_id = profile.id;
  }
  if (draft.options_revision !== options.revision) {
    patch.options_revision = options.revision;
  }
  return patch;
}

export function defaultWorkerModelBindingsPatch(
  draft: WorkerSpecDraft,
  workerType: WorkerTypeOption,
  resources: EffectiveResource[],
  providers: ProviderDefinition[],
): Partial<WorkerSpecDraft> {
  const patch: Partial<WorkerSpecDraft> = {};
  const primaryRequirement = primaryModelRequirement(workerType);
  const primaryResources = primaryRequirement
    ? compatibleWorkerModelResources(resources, providers, primaryRequirement)
    : [];
  const primaryId = selectedResourceId(
    primaryResources,
    draft.model_resource_id,
  );
  const nextPrimaryId = primaryRequirement ? primaryId : 0;
  if (draft.model_resource_id !== nextPrimaryId) {
    patch.model_resource_id = nextPrimaryId;
  }

  const nextTools: Record<string, number> = {};
  for (const requirement of workerType.tool_model_requirements) {
    const candidates = compatibleWorkerModelResources(
      resources,
      providers,
      toolModelRequirement(requirement),
    );
    const selected = selectedResourceId(
      candidates,
      draft.tool_model_resource_ids[requirement.role] ?? 0,
    );
    if (selected > 0) nextTools[requirement.role] = selected;
  }
  if (!sameRecord(draft.tool_model_resource_ids, nextTools)) {
    patch.tool_model_resource_ids = nextTools;
  }
  return patch;
}

function selectedResourceId(
  resources: EffectiveResource[],
  currentId: number,
): number {
  if (resources.some((item) => item.resource?.id === currentId)) return currentId;
  return resources[0]?.resource?.id ?? 0;
}

function sameRecord(
  left: Record<string, number>,
  right: Record<string, number>,
): boolean {
  const keys = Object.keys(left);
  return keys.length === Object.keys(right).length &&
    keys.every((key) => left[key] === right[key]);
}

function selectedOrFirst<T extends { selectable: boolean }>(
  options: T[],
  selected: string,
  value: (option: T) => string,
): T | undefined {
  const current = options.find((option) => value(option) === selected);
  if (current?.selectable) return current;
  return options.find((option) => option.selectable);
}
