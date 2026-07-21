import {
  ArtifactStatus,
  type ArtifactDescriptor,
  type ArtifactRepresentation,
} from "@agent-cloud/proto/agent_workbench/v2/artifact_pb";

export function selectArtifactRepresentation(
  descriptor: ArtifactDescriptor,
  requestedId?: string,
): ArtifactRepresentation | undefined {
  if (requestedId) {
    return descriptor.representations.find(
      (representation) => representation.representationId === requestedId,
    );
  }
  const manifestId = manifestRepresentationId(descriptor);
  if (manifestId) {
    return descriptor.representations.find(
      (representation) => representation.representationId === manifestId,
    );
  }
  return (
    descriptor.representations.find(
      (representation) =>
        representation.status === ArtifactStatus.READY &&
        representation.representationId === "preview-pdf",
    ) ??
    descriptor.representations.find(
      (representation) =>
        representation.status === ArtifactStatus.READY &&
        representation.role === "primary",
    ) ??
    descriptor.representations.find(
      (representation) => representation.status === ArtifactStatus.READY,
    ) ??
    descriptor.representations[0]
  );
}

export function extensionSchemaVersion(
  descriptor: ArtifactDescriptor,
): string | undefined {
  const manifest = descriptor.manifest?.manifest;
  return manifest?.case === "extension" ? manifest.value.schemaVersion : undefined;
}

function manifestRepresentationId(
  descriptor: ArtifactDescriptor,
): string | undefined {
  const manifest = descriptor.manifest?.manifest;
  if (!manifest?.case) return undefined;
  if (manifest.case === "imageEdit") {
    return (
      manifest.value.resultRepresentationId ??
      manifest.value.candidateRepresentationIds[0] ??
      manifest.value.sourceRepresentationId
    );
  }
  if (manifest.case === "video") {
    return (
      manifest.value.playableRepresentationId ??
      manifest.value.originalRepresentationId ??
      manifest.value.posterRepresentationId
    );
  }
  if (manifest.case !== "presentation") return undefined;
  const selected = manifest.value.versions.find(
    (version) => version.versionId === manifest.value.selectedVersionId,
  );
  const revision = descriptor.revisions.find(
    (candidate) =>
      candidate.revision === (selected?.revision ?? manifest.value.deckRevision),
  );
  return (
    revision?.representationIds[0] ??
    manifest.value.slides.find((slide) => slide.pageRepresentationId)
      ?.pageRepresentationId
  );
}
