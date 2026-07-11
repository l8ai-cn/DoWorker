import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";

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

export function defaultModelPatch(
  draft: WorkerSpecDraft,
  resources: EffectiveResource[],
): Partial<WorkerSpecDraft> {
  if (draft.model_resource_id > 0) return {};
  const selected = resources.find((item) => item.selectable && item.resource?.id);
  return selected?.resource?.id
    ? { model_resource_id: selected.resource.id }
    : {};
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
