export interface ArtifactActionCommand<
  TActionType extends string = string,
  TPayload = unknown,
> {
  readonly artifactId: string;
  readonly representationId: string;
  readonly baseRevision: bigint;
  readonly clientActionId: string;
  readonly actionType: TActionType;
  readonly payload: TPayload;
}

export function artifactAction<
  const TActionType extends string,
  TPayload,
>(
  command: ArtifactActionCommand<TActionType, TPayload>,
): ArtifactActionCommand<TActionType, TPayload> {
  return command;
}

export class ArtifactRevisionConflictError extends Error {
  readonly code = "artifact_revision_conflict";
  readonly baseRevision: bigint;
  readonly currentRevision: bigint;

  constructor(
    baseRevision: bigint,
    currentRevision: bigint,
  ) {
    super(
      `artifact_revision_conflict: base_revision=${baseRevision} current_revision=${currentRevision}`,
    );
    this.name = "ArtifactRevisionConflictError";
    this.baseRevision = baseRevision;
    this.currentRevision = currentRevision;
  }
}

export class ArtifactActionIdConflictError extends Error {
  readonly code = "artifact_action_id_conflict";
  readonly clientActionId: string;

  constructor(clientActionId: string) {
    super(`artifact_action_id_conflict: client_action_id=${clientActionId}`);
    this.name = "ArtifactActionIdConflictError";
    this.clientActionId = clientActionId;
  }
}
