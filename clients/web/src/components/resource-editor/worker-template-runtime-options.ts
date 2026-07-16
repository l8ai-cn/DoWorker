import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import type { WorkerTemplateDraft } from "./resource-editor-types";

interface RuntimeOption {
  value: string;
  label: string;
  selectable: boolean;
  blockingReason: string;
}

export function synchronizeWorkerTemplateRuntime(
  draft: WorkerTemplateDraft,
  options: WorkerCreateOptions,
): WorkerTemplateDraft | undefined {
  const type = options.worker_types.find(
    (option) => option.slug === draft.spec.workerType,
  ) ?? (draft.spec.workerType ? undefined : firstSelectable(options.worker_types));
  if (!type) {
    return draft.spec.optionsRevision === options.revision ? undefined : {
      ...draft,
      spec: { ...draft.spec, optionsRevision: options.revision },
    };
  }
  const image = selectedOrFirstImage(
    options,
    type.slug,
    draft.spec.runtime.runtimeImageId,
  );
  const interactionMode = supportedInteractionMode(
    type.supported_interaction_modes,
    draft.spec.typeConfig.interactionMode,
  );
  if (
    draft.spec.workerType === type.slug &&
    draft.spec.optionsRevision === options.revision &&
    draft.spec.runtime.runtimeImageId === (image?.id ?? 0) &&
    draft.spec.typeConfig.schemaVersion === type.schema_version &&
    draft.spec.typeConfig.interactionMode === interactionMode
  ) return undefined;
  return withRuntimeChoice(draft, options.revision, type, image?.id ?? 0);
}

export function selectWorkerTemplateType(
  draft: WorkerTemplateDraft,
  options: WorkerCreateOptions,
  slug: string,
): WorkerTemplateDraft | undefined {
  const type = options.worker_types.find(
    (option) => option.slug === slug && option.selectable,
  );
  if (!type) return undefined;
  const image = firstCompatibleImage(options, slug);
  return withRuntimeChoice(draft, options.revision, type, image?.id ?? 0);
}

export function workerTemplateTypeOptions(
  options: WorkerCreateOptions,
): RuntimeOption[] {
  return options.worker_types.map((option) => ({
    value: option.slug,
    label: option.name,
    selectable: option.selectable,
    blockingReason: option.blocking_reason,
  }));
}

export function workerTemplateRuntimeImageOptions(
  options: WorkerCreateOptions,
  workerType: string,
): RuntimeOption[] {
  return options.runtime_images
    .filter((option) => option.worker_type_slugs.includes(workerType))
    .map((option) => ({
      value: String(option.id),
      label: option.name,
      selectable: option.selectable,
      blockingReason: option.blocking_reason,
    }));
}

function withRuntimeChoice(
  draft: WorkerTemplateDraft,
  optionsRevision: string,
  type: WorkerCreateOptions["worker_types"][number],
  runtimeImageId: number,
): WorkerTemplateDraft {
  return {
    ...draft,
    spec: {
      ...draft.spec,
      optionsRevision,
      workerType: type.slug,
      runtime: { ...draft.spec.runtime, runtimeImageId },
      typeConfig: {
        ...draft.spec.typeConfig,
        schemaVersion: type.schema_version,
        interactionMode: supportedInteractionMode(
          type.supported_interaction_modes,
          draft.spec.typeConfig.interactionMode,
        ),
      },
    },
  };
}

function selectedOrFirstImage(
  options: WorkerCreateOptions,
  workerType: string,
  selectedID: number,
) {
  return options.runtime_images.find((option) =>
    option.id === selectedID && option.worker_type_slugs.includes(workerType),
  ) ?? (selectedID ? undefined : firstCompatibleImage(options, workerType));
}

function firstCompatibleImage(options: WorkerCreateOptions, workerType: string) {
  return options.runtime_images.find(
    (option) => option.selectable && option.worker_type_slugs.includes(workerType),
  );
}

function firstSelectable<T extends { selectable: boolean }>(
  options: T[],
): T | undefined {
  return options.find((option) => option.selectable);
}

function supportedInteractionMode(modes: string[], current: string): string {
  return modes.includes(current) ? current : modes[0] ?? current;
}
