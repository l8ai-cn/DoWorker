import {
  assertResourceDraftIdentity,
  type ResourceIdentity,
} from "./resource-draft-identity";
import { createResourceDraft } from "./resource-draft-factory";
import { createResourceDraftState } from "./resource-draft-reducer";
import type { ResourceDraft, ResourceEditorKind } from "./resource-editor-types";

export function createResourceEditorInitialState(
  orgSlug: string,
  kind: ResourceEditorKind,
  initialDraft: ResourceDraft | undefined,
  lockedIdentity: ResourceIdentity | undefined,
) {
  if (initialDraft && initialDraft.kind !== kind) {
    throw new Error(`Expected ${kind} draft, received ${initialDraft.kind}.`);
  }
  const draft = initialDraft ?? createResourceDraft(kind, orgSlug);
  assertResourceDraftIdentity(draft, lockedIdentity);
  return createResourceDraftState(draft);
}
