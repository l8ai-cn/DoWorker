import type {
  ResourceDraft,
  ResourceEditorKind,
} from "./resource-editor-types";

export interface ResourceIdentity {
  apiVersion: string;
  kind: ResourceEditorKind;
  namespace: string;
  name: string;
}

export function resourceDraftIdentity(
  draft: ResourceDraft,
): ResourceIdentity {
  return {
    apiVersion: draft.apiVersion,
    kind: draft.kind,
    namespace: draft.metadata.namespace,
    name: draft.metadata.name,
  };
}

export function assertResourceDraftIdentity(
  draft: ResourceDraft,
  lockedIdentity?: ResourceIdentity,
): void {
  if (!lockedIdentity) return;
  const current = resourceDraftIdentity(draft);
  if (
    current.apiVersion !== lockedIdentity.apiVersion ||
    current.kind !== lockedIdentity.kind ||
    current.namespace !== lockedIdentity.namespace ||
    current.name !== lockedIdentity.name
  ) {
    throw new Error(
      "Resource identity cannot change when creating a revision.",
    );
  }
}
