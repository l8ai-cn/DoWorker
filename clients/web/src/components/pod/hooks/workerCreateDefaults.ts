import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { EnvBundleSummary } from "@/lib/api";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import { compatibleToolModelResources } from "../CreatePodForm/workerModelResources";

export function defaultWorkerDraftPatch(
  draft: WorkerSpecDraft,
  options: WorkerCreateOptions,
  preferredWorkerType?: string,
): Partial<WorkerSpecDraft> {
  const patch: Partial<WorkerSpecDraft> = {};
  const workerType = selectedOrFirst(
    options.worker_types,
    draft.worker_type_slug || preferredWorkerType || "codex-cli",
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

export function defaultToolModelPatch(
  draft: WorkerSpecDraft,
  workerType: WorkerCreateOptions["worker_types"][number] | undefined,
  resources: EffectiveResource[],
): Partial<WorkerSpecDraft> {
  if (!workerType?.tool_model_requirements.length) return {};
  const selected = { ...draft.tool_model_resource_ids };
  for (const requirement of workerType.tool_model_requirements) {
    if ((selected[requirement.role] ?? 0) > 0) continue;
    const resource = compatibleToolModelResources(requirement, resources)
      .find((item) => item.resource?.id);
    if (resource?.resource?.id) selected[requirement.role] = resource.resource.id;
  }
  return changedMap(draft.tool_model_resource_ids, selected)
    ? { tool_model_resource_ids: selected }
    : {};
}

export function defaultConfigDocumentPatch(
  draft: WorkerSpecDraft,
  workerType: WorkerCreateOptions["worker_types"][number] | undefined,
  bundles: EnvBundleSummary[],
): Partial<WorkerSpecDraft> {
  const required = workerType?.config_document_requirements
    .filter((requirement) => requirement.required) ?? [];
  if (required.length === 0) return {};
  const bindings = [...draft.config_document_bindings];
  for (const requirement of required) {
    if (bindings.some((binding) => binding.document_id === requirement.document_id)) {
      continue;
    }
    const bundle = bundles.find(
      (item) => item.kind === "config" && item.agent_slug === workerType?.slug,
    );
    if (bundle) {
      bindings.push({
        document_id: requirement.document_id,
        config_bundle_id: bundle.id,
      });
    }
  }
  return bindings.length === draft.config_document_bindings.length
    ? {}
    : { config_document_bindings: bindings };
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

function changedMap(left: Record<string, number>, right: Record<string, number>): boolean {
  const keys = new Set([...Object.keys(left), ...Object.keys(right)]);
  for (const key of keys) {
    if ((left[key] ?? 0) !== (right[key] ?? 0)) return true;
  }
  return false;
}
