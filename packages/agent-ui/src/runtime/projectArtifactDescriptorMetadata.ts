import {
  ArtifactStatus,
  type ArtifactGrant,
  type ArtifactRepresentation,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";

import type {
  AgentArtifactGrant,
  AgentArtifactItem,
  AgentArtifactRepresentation,
} from "../agentArtifactContracts";

export function projectRepresentations(
  representations: readonly ArtifactRepresentation[],
): AgentArtifactRepresentation[] {
  return representations.map((representation) => ({
    representationId: representation.representationId,
    revision: representation.revision,
    mediaType: representation.mediaType,
    ...(representation.role ? { role: representation.role } : {}),
    ...(representation.filename ? { filename: representation.filename } : {}),
    status: representationStatus(representation.status),
    ...(representation.byteSize !== undefined
      ? { byteSize: representation.byteSize }
      : {}),
    ...(representation.dimensions
      ? { dimensions: { ...representation.dimensions } }
      : {}),
    ...(representation.durationMillis !== undefined
      ? { durationMillis: representation.durationMillis }
      : {}),
    ...(representation.digest ? { digest: representation.digest } : {}),
  }));
}

export function projectGrants(
  grants: readonly ArtifactGrant[],
): AgentArtifactGrant[] {
  return grants.map((grant) => ({
    grantId: grant.grantId,
    representationIds: [...grant.representationIds],
    actions: [...grant.actions],
    ...(grant.issuer ? { issuer: grant.issuer } : {}),
    ...(grant.subject ? { subject: grant.subject } : {}),
    ...(grant.minimumRevision !== undefined
      ? { minimumRevision: grant.minimumRevision }
      : {}),
    ...(grant.maximumRevision !== undefined
      ? { maximumRevision: grant.maximumRevision }
      : {}),
    ...(grant.issuedAt ? { issuedAt: grant.issuedAt } : {}),
    ...(grant.expiresAt ? { expiresAt: grant.expiresAt } : {}),
  }));
}

export function projectGrantActions(
  grants: readonly AgentArtifactGrant[],
  revision: bigint,
): string[] {
  return [
    ...new Set(
      grants
        .filter((grant) => revisionApplies(grant, revision))
        .flatMap((grant) => grant.actions),
    ),
  ];
}

export function projectArtifactStatus(
  status: ArtifactStatus,
): AgentArtifactItem["status"] {
  if (status === ArtifactStatus.QUEUED) return "queued";
  if (status === ArtifactStatus.PROCESSING) return "processing";
  if (status === ArtifactStatus.READY) return "completed";
  return "failed";
}

function revisionApplies(
  grant: AgentArtifactGrant,
  revision: bigint,
): boolean {
  if (grant.minimumRevision !== undefined && revision < grant.minimumRevision) {
    return false;
  }
  return grant.maximumRevision === undefined || revision <= grant.maximumRevision;
}

function representationStatus(
  status: ArtifactStatus,
): AgentArtifactRepresentation["status"] {
  if (status === ArtifactStatus.QUEUED) return "queued";
  if (status === ArtifactStatus.PROCESSING) return "processing";
  if (status === ArtifactStatus.READY) return "ready";
  if (status === ArtifactStatus.FAILED) return "failed";
  if (status === ArtifactStatus.DELETED) return "deleted";
  return "unknown";
}
