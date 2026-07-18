import type { ResourceIdentity } from "./resource-draft-identity";
import type { ResourceEditorKind } from "./resource-editor-types";

export function resourceEditorInstanceIdentity(
  orgSlug: string,
  kind: ResourceEditorKind,
  sessionKey: string | undefined,
  lockedIdentity: ResourceIdentity | undefined,
): string {
  return JSON.stringify([
    orgSlug,
    kind,
    sessionKey ?? null,
    lockedIdentity
      ? [
          lockedIdentity.apiVersion,
          lockedIdentity.kind,
          lockedIdentity.namespace,
          lockedIdentity.name,
        ]
      : null,
  ]);
}
