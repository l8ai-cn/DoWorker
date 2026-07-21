import type { ArtifactDescriptor } from "@agent-cloud/proto/agent_workbench/v2/artifact_pb";

import type { AgentActivityItem, AgentArtifactItem } from "../contracts";
import {
  projectArtifactStatus,
  projectGrantActions,
  projectGrants,
  projectRepresentations,
} from "./projectArtifactDescriptorMetadata";
import { projectManifest } from "./projectArtifactManifest";
import {
  extensionSchemaVersion,
  selectArtifactRepresentation,
} from "./projectArtifactRepresentationSelection";

export type ArtifactCatalog = ReadonlyMap<string, ArtifactDescriptor>;

export interface ArtifactProjectionReference {
  artifactId: string;
  filename?: string;
  mediaType?: string;
  representationId?: string;
  revision?: bigint;
  role?: string;
}

export function createArtifactCatalog(
  artifacts: readonly ArtifactDescriptor[],
): ArtifactCatalog {
  return new Map(artifacts.map((artifact) => [artifact.artifactId, artifact]));
}

export function projectArtifactDescriptor(
  descriptor: ArtifactDescriptor,
  id: string,
  options: {
    representationId?: string;
    role?: string;
    schemaVersion?: string;
  } = {},
): AgentArtifactItem {
  const representation = selectArtifactRepresentation(
    descriptor,
    options.representationId,
  );
  const representations = projectRepresentations(descriptor.representations);
  const grants = projectGrants(descriptor.grants);
  const selectedRepresentationId = representation?.representationId ?? options.representationId ?? null;
  return {
    actions: projectGrantActions(grants, descriptor.revision),
    id,
    kind: "artifact",
    artifactId: descriptor.artifactId,
    filename:
      representation?.representationId === "preview-pdf"
        ? descriptor.filename
        : representation?.filename || descriptor.filename,
    grants,
    manifest: projectManifest(descriptor.manifest),
    mimeType: representation?.mediaType || descriptor.mediaType || null,
    provenance: {
      ...(currentRevisionToolExecutionId(descriptor)
        ? {
            publicationToolExecutionId:
              currentRevisionToolExecutionId(descriptor),
          }
        : {}),
      ...(currentRevisionCommandId(descriptor)
        ? { commandId: currentRevisionCommandId(descriptor) }
        : {}),
      ...(descriptor.provenance?.producerId
        ? { producerId: descriptor.provenance.producerId }
        : {}),
      ...(descriptor.provenance?.producerNamespace
        ? { producerNamespace: descriptor.provenance.producerNamespace }
        : {}),
      ...(descriptor.provenance?.producerType
        ? { producerType: descriptor.provenance.producerType }
        : {}),
    },
    representations,
    revision: descriptor.revision,
    role: options.role || descriptor.role || representation?.role || "artifact",
    schemaVersion:
      options.schemaVersion || extensionSchemaVersion(descriptor) || "1",
    selectedRepresentationId,
    status: projectArtifactStatus(representation?.status ?? descriptor.status),
  };
}

function currentRevisionToolExecutionId(
  descriptor: ArtifactDescriptor,
): string | undefined {
  return descriptor.revisions.find(
    (revision) => revision.revision === descriptor.revision,
  )?.provenance?.toolExecutionId;
}

function currentRevisionCommandId(
  descriptor: ArtifactDescriptor,
): string | undefined {
  return descriptor.revisions.find(
    (revision) => revision.revision === descriptor.revision,
  )?.provenance?.commandId;
}

export function projectArtifactReference(
  reference: ArtifactProjectionReference | undefined,
  id: string,
  catalog: ArtifactCatalog,
  options: { label?: string; schemaVersion?: string } = {},
): AgentArtifactItem | AgentActivityItem {
  if (!reference) return missingArtifact(id, "artifact reference is missing");
  const descriptor = catalog.get(reference.artifactId);
  if (!descriptor) {
    return missingArtifact(
      id,
      `artifactId=${reference.artifactId}; representationId=${reference.representationId ?? "unspecified"}`,
    );
  }
  if (
    reference.revision !== undefined &&
    descriptor.revision !== reference.revision
  ) {
    return missingArtifact(
      id,
      `artifactId=${reference.artifactId}; revision=${reference.revision.toString()}; currentRevision=${descriptor.revision.toString()}`,
    );
  }
  if (
    reference.representationId &&
    !descriptor.representations.some(
      (representation) =>
        representation.representationId === reference.representationId,
    )
  ) {
    return missingArtifact(
      id,
      `artifactId=${reference.artifactId}; representationId=${reference.representationId}`,
    );
  }
  return projectArtifactDescriptor(descriptor, id, {
    representationId: reference.representationId,
    role: reference.role,
    schemaVersion: options.schemaVersion,
  });
}

function missingArtifact(id: string, detail: string): AgentActivityItem {
  return {
    id,
    kind: "system",
    title: "Unsupported artifact",
    detail,
    status: "failed",
  };
}
