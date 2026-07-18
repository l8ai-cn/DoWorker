import type { AgentArtifactItem } from "./agentArtifactContracts";
import { artifactPresentation } from "./artifactPresentation";

const SHA256_DIGEST = /^sha256:[a-f0-9]{64}$/;

export function isUserVisibleArtifact(item: AgentArtifactItem): boolean {
  if (!isVideoArtifact(item)) return true;
  if (item.status === "completed") return isVerifiedReadyVideoArtifact(item);
  if (item.manifest?.kind !== "video") return false;
  if (item.status === "failed") return item.manifest.stage === "failed";
  if (item.status === "queued") return item.manifest.stage === "queued";
  return (
    item.manifest.stage === "rendering" ||
    item.manifest.stage === "transcoding"
  );
}

export function isVerifiedReadyVideoArtifact(
  item: AgentArtifactItem,
): boolean {
  if (item.status !== "completed" || item.manifest?.kind !== "video") {
    return false;
  }
  if (
    item.manifest.stage !== "ready" ||
    !item.manifest.playableRepresentationId ||
    !item.provenance?.publicationToolExecutionId
  ) {
    return false;
  }
  const playableRepresentationId = item.manifest.playableRepresentationId;
  const playable = item.representations.find(
    (representation) =>
      representation.representationId === playableRepresentationId,
  );
  return Boolean(
    playable &&
      playable.status === "ready" &&
      playable.mediaType.startsWith("video/") &&
      playable.byteSize !== undefined &&
      playable.byteSize > BigInt(0) &&
      playable.digest &&
      SHA256_DIGEST.test(playable.digest),
  );
}

export function isVideoArtifact(item: AgentArtifactItem): boolean {
  return (
    item.manifest?.kind === "video" ||
    artifactPresentation(item.mimeType, item.filename).kind === "video"
  );
}
