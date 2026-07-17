import type { AgentArtifactItem } from "../agentArtifactContracts";
import { isSafeArtifactImage } from "../safeArtifactImage";
import type { ContentRendererKey } from "./rendererKeys";

export function artifactRendererKey(
  item: AgentArtifactItem,
): ContentRendererKey | null {
  if (item.manifest?.kind === "image_edit") {
    return manifestKey(item, "image_edit");
  }
  if (item.manifest?.kind === "video") {
    return manifestKey(item, "video");
  }
  if (
    item.manifest?.kind === "presentation" &&
    hasPresentationPages(item)
  ) {
    return manifestKey(item, "presentation");
  }
  if (!item.mimeType) return null;
  return {
    blockKind: "artifact",
    mediaType: item.mimeType,
    role: item.role,
    schemaVersion: item.schemaVersion,
  };
}

export function artifactUnsupportedReason(
  item: AgentArtifactItem,
): string | null {
  if (item.role === "image_edit" && item.manifest?.kind !== "image_edit") {
    return "image_edit_manifest_missing";
  }
  return null;
}

function manifestKey(
  item: AgentArtifactItem,
  manifestType: string,
): ContentRendererKey {
  return {
    blockKind: "artifact",
    manifestType,
    schemaVersion: item.schemaVersion,
  };
}

function hasPresentationPages(item: AgentArtifactItem): boolean {
  if (item.manifest?.kind !== "presentation") return false;
  const representations = new Map(
    item.representations.map((representation) => [
      representation.representationId,
      representation,
    ]),
  );
  return item.manifest.slides.some((slide) => {
    if (!slide.pageRepresentationId) return false;
    const representation = representations.get(slide.pageRepresentationId);
    return (
      representation?.status === "ready" &&
      isSafeArtifactImage(representation.mediaType)
    );
  });
}
