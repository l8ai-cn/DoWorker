import {
  ArtifactActionIdConflictError,
  ArtifactRevisionConflictError,
  type ArtifactActionCommand,
} from "./artifactAction";
import type {
  ArtifactDescriptor,
  ArtifactPayload,
  ArtifactRuntime,
} from "./ArtifactRuntime";

export class ArtifactController<
  TRepresentationPayload = ArtifactPayload,
  TActionResult = unknown,
> {
  private readonly actionResults = new Map<
    string,
    {
      command: ArtifactActionCommand;
      result: Promise<TActionResult>;
    }
  >();
  private readonly descriptors = new Map<string, ArtifactDescriptor>();
  private readonly representations = new Map<
    string,
    Promise<TRepresentationPayload>
  >();

  constructor(
    private readonly runtime: ArtifactRuntime<
      TRepresentationPayload,
      TActionResult
    >,
  ) {}

  updateDescriptor(descriptor: ArtifactDescriptor): void {
    this.descriptors.set(
      descriptorKey(descriptor.artifactId, descriptor.representationId),
      descriptor,
    );
  }

  getDescriptor(
    artifactId: string,
    representationId: string,
  ): ArtifactDescriptor | undefined {
    return this.descriptors.get(descriptorKey(artifactId, representationId));
  }

  loadRepresentation(
    artifactId: string,
    representationId: string,
    revision: bigint,
  ): Promise<TRepresentationPayload> {
    const key = representationKey(artifactId, representationId, revision);
    const cached = this.representations.get(key);
    if (cached) return cached;

    const pending = this.runtime.loadRepresentation(
      artifactId,
      representationId,
      revision,
    );
    this.representations.set(key, pending);
    void pending.catch(() => {
      if (this.representations.get(key) === pending) {
        this.representations.delete(key);
      }
    });
    return pending;
  }

  execute(command: ArtifactActionCommand): Promise<TActionResult> {
    const cached = this.actionResults.get(command.clientActionId);
    if (cached) {
      if (!sameAction(cached.command, command)) {
        return Promise.reject(
          new ArtifactActionIdConflictError(command.clientActionId),
        );
      }
      return cached.result;
    }

    const descriptor = this.getDescriptor(
      command.artifactId,
      command.representationId,
    );
    if (!descriptor) {
      return Promise.reject(
        new Error(
          `artifact_descriptor_missing: artifact_id=${command.artifactId} representation_id=${command.representationId}`,
        ),
      );
    }
    if (descriptor.revision !== command.baseRevision) {
      return Promise.reject(
        new ArtifactRevisionConflictError(
          command.baseRevision,
          descriptor.revision,
        ),
      );
    }

    const pending = this.runtime.executeAction(command);
    this.actionResults.set(command.clientActionId, {
      command,
      result: pending,
    });
    return pending;
  }
}

function descriptorKey(
  artifactId: string,
  representationId: string,
): string {
  return JSON.stringify([artifactId, representationId]);
}

function representationKey(
  artifactId: string,
  representationId: string,
  revision: bigint,
): string {
  return JSON.stringify([artifactId, representationId, revision.toString()]);
}

function sameAction(
  left: ArtifactActionCommand,
  right: ArtifactActionCommand,
): boolean {
  return (
    left.artifactId === right.artifactId &&
    left.representationId === right.representationId &&
    left.baseRevision === right.baseRevision &&
    left.clientActionId === right.clientActionId &&
    left.actionType === right.actionType &&
    sameValue(left.payload, right.payload)
  );
}

function sameValue(left: unknown, right: unknown): boolean {
  if (Object.is(left, right)) return true;
  if (
    typeof left !== "object" ||
    typeof right !== "object" ||
    left === null ||
    right === null
  ) {
    return false;
  }
  if (Array.isArray(left) || Array.isArray(right)) {
    return (
      Array.isArray(left) &&
      Array.isArray(right) &&
      left.length === right.length &&
      left.every((value, index) => sameValue(value, right[index]))
    );
  }
  if (
    Object.getPrototypeOf(left) !== Object.getPrototypeOf(right) ||
    (Object.getPrototypeOf(left) !== Object.prototype &&
      Object.getPrototypeOf(left) !== null)
  ) {
    return false;
  }
  const leftRecord = left as Record<string, unknown>;
  const rightRecord = right as Record<string, unknown>;
  const leftKeys = Object.keys(leftRecord);
  const rightKeys = Object.keys(rightRecord);
  return (
    leftKeys.length === rightKeys.length &&
    leftKeys.every(
      (key) =>
        Object.hasOwn(rightRecord, key) &&
        sameValue(leftRecord[key], rightRecord[key]),
    )
  );
}
