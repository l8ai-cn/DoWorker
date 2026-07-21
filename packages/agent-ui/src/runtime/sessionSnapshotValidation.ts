import { ArtifactStatus } from "@agent-cloud/proto/agent_workbench/v2/artifact_pb";
import {
  CommandReceiptState,
  PermissionDecision,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import type { SessionSnapshot } from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  SessionResourceStatus,
  SessionStatus,
} from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import { AgentSessionReductionError } from "./agentSessionState";

export function validateSessionSnapshot(snapshot: SessionSnapshot): void {
  if (
    !snapshot.sessionId ||
    !snapshot.streamEpoch ||
    snapshot.status === SessionStatus.UNSPECIFIED
  ) {
    throw new AgentSessionReductionError("snapshot_identity_missing");
  }
  if (
    snapshot.revision > snapshot.latestSequence ||
    (snapshot.revision === BigInt(0)) !==
      (snapshot.latestSequence === BigInt(0))
  ) {
    throw new AgentSessionReductionError("snapshot_position_invalid");
  }

  validateHistory(snapshot);
  validateReceipts(snapshot);
  validateGrants(snapshot);
  validatePermissions(snapshot);
  validateResources(snapshot);
  validateArtifacts(snapshot);
}

function validateHistory(snapshot: SessionSnapshot): void {
  const itemIds = new Set<string>();
  const sequences = new Set<bigint>();
  for (const item of snapshot.history) {
    const envelope = item.envelope;
    if (
      !envelope ||
      !item.content ||
      item.content.content.case === undefined ||
      !envelope.itemId ||
      !envelope.createdAt ||
      envelope.sessionId !== snapshot.sessionId ||
      envelope.streamEpoch !== snapshot.streamEpoch ||
      envelope.revision > snapshot.revision ||
      envelope.sequence > snapshot.latestSequence ||
      itemIds.has(envelope.itemId) ||
      sequences.has(envelope.sequence)
    ) {
      throw new AgentSessionReductionError("snapshot_history_invalid");
    }
    itemIds.add(envelope.itemId);
    sequences.add(envelope.sequence);
  }
}

function validateReceipts(snapshot: SessionSnapshot): void {
  const commandIds = new Set<string>();
  for (const receipt of snapshot.commandReceipts) {
    if (
      receipt.sessionId !== snapshot.sessionId ||
      !receipt.commandId ||
      !receipt.payloadDigest ||
      receipt.state === CommandReceiptState.UNSPECIFIED ||
      commandIds.has(receipt.commandId) ||
      (receipt.resultingRevision !== undefined &&
        receipt.resultingRevision > snapshot.revision)
    ) {
      throw new AgentSessionReductionError("snapshot_receipt_invalid");
    }
    commandIds.add(receipt.commandId);
  }
}

function validateGrants(snapshot: SessionSnapshot): void {
  const grantIds = new Set<string>();
  for (const grant of snapshot.grants) {
    if (
      !grant.grantId ||
      grant.sessionId !== snapshot.sessionId ||
      grantIds.has(grant.grantId)
    ) {
      throw new AgentSessionReductionError("snapshot_grant_invalid");
    }
    grantIds.add(grant.grantId);
  }
}

function validatePermissions(snapshot: SessionSnapshot): void {
  validateUniqueIds(
    snapshot.permissionRequests,
    "permissionRequestId",
    (request) => {
      if (request.state === PermissionRequestState.UNSPECIFIED) return false;
      if (request.state === PermissionRequestState.RESOLVED) {
        return (
          request.resolution?.permissionRequestId === request.permissionRequestId &&
          request.resolution.decision !== PermissionDecision.UNSPECIFIED
        );
      }
      if (request.state === PermissionRequestState.PENDING) {
        return request.request.case !== undefined && request.resolution === undefined;
      }
      return request.resolution === undefined;
    },
    "snapshot_permission_invalid",
  );
}

function validateResources(snapshot: SessionSnapshot): void {
  validateUniqueIds(
    snapshot.resources,
    "resourceId",
    (resource) =>
      resource.status !== SessionResourceStatus.UNSPECIFIED &&
      resource.resource.case !== undefined,
    "snapshot_resource_invalid",
  );
}

function validateArtifacts(snapshot: SessionSnapshot): void {
  validateUniqueIds(
    snapshot.artifacts,
    "artifactId",
    (artifact) =>
      artifact.status !== ArtifactStatus.UNSPECIFIED &&
      artifact.revision > BigInt(0) &&
      Boolean(artifact.filename && artifact.mediaType),
    "snapshot_artifact_invalid",
  );
}

function validateUniqueIds<T, K extends keyof T>(
  values: readonly T[],
  key: K,
  isValid: (value: T) => boolean,
  errorCode: string,
): void {
  const ids = new Set<T[K]>();
  for (const value of values) {
    const id = value[key];
    if (!id || ids.has(id) || !isValid(value)) {
      throw new AgentSessionReductionError(errorCode);
    }
    ids.add(id);
  }
}
