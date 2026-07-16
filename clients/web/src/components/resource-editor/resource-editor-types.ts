import type {
  DefinitionResourceDraft,
  DefinitionResourceKind,
} from "./resource-definition-draft-types";
import type {
  ResourceBindingDraft,
  ResourceBindingKind,
} from "./resource-binding-draft-types";
import { isResourceBindingKind } from "./resource-binding-draft-types";
import type {
  WorkerResourceDraft,
  WorkerResourceKind,
  WorkerTemplateDraft,
} from "./worker-resource-draft-types";

export * from "./resource-binding-draft-types";
export * from "./resource-definition-draft-types";
export * from "./resource-manifest-types";
export * from "./worker-resource-draft-types";

export type ResourceEditorKind =
  | WorkerResourceKind
  | DefinitionResourceKind
  | ResourceBindingKind;

export type ResourceDraft =
  | WorkerResourceDraft
  | DefinitionResourceDraft
  | ResourceBindingDraft;

export function isWorkerTemplateDraft(
  draft: ResourceDraft,
): draft is WorkerTemplateDraft {
  return draft.kind === "WorkerTemplate";
}

export function isResourceBindingDraft(
  draft: ResourceDraft,
): draft is ResourceBindingDraft {
  return isResourceBindingKind(draft.kind);
}
