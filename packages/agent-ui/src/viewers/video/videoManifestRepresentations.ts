import type {
  AgentArtifactRepresentation,
  AgentVideoManifest,
} from "../../agentArtifactContracts";
import { isSafeArtifactImage } from "../../safeArtifactImage";
import type { ArtifactRepresentationUrlState } from "../../useArtifactRepresentationUrls";
import type { VideoArtifactVersion } from "./VideoArtifactViewer";

export function videoRepresentations(
  representations: readonly AgentArtifactRepresentation[],
  manifest: AgentVideoManifest,
) {
  const ids = new Set(
    [
      manifest.playableRepresentationId,
      manifest.originalRepresentationId,
      ...manifest.derivativeRepresentationIds,
    ].filter((value): value is string => Boolean(value)),
  );
  return representations.filter(
    (representation) =>
      ids.has(representation.representationId) &&
      representation.status === "ready",
  );
}

export function videoPosterRepresentation(
  representations: readonly AgentArtifactRepresentation[],
  manifest: AgentVideoManifest,
) {
  return representations.find(
    (representation) =>
      representation.representationId === manifest.posterRepresentationId &&
      representation.status === "ready" &&
      isSafeArtifactImage(representation.mediaType),
  );
}

export function videoVersions(
  representations: readonly AgentArtifactRepresentation[],
  resources: Readonly<Record<string, ArtifactRepresentationUrlState>>,
): VideoArtifactVersion[] {
  return representations.map((representation, index) => {
    const resource = resources[representation.representationId];
    return {
      id: representation.representationId,
      label: representation.role || `版本 ${index + 1}`,
      filename: representation.filename,
      mimeType: representation.mediaType,
      durationSeconds:
        representation.durationMillis === undefined
          ? undefined
          : Number(representation.durationMillis) / 1000,
      src: resource?.status === "ready" ? resource.url : undefined,
    };
  });
}
