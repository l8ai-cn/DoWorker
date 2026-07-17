import type {
  PresentationActionPayload,
  PresentationArtifactAction,
  PresentationGrant,
} from "./presentationContracts";

export interface PresentationActionContext {
  readonly actionSchemaVersion: string;
  readonly artifactId: string;
  readonly baseRevision: bigint;
  readonly representationId?: string;
}

export function createPresentationAction(
  context: PresentationActionContext,
  actionType: PresentationGrant,
  payload: PresentationActionPayload,
): PresentationArtifactAction {
  return {
    actionSchemaVersion: context.actionSchemaVersion,
    actionType,
    artifactId: context.artifactId,
    baseRevision: context.baseRevision,
    commandId: crypto.randomUUID(),
    payload,
    ...(context.representationId
      ? { representationId: context.representationId }
      : {}),
  };
}
