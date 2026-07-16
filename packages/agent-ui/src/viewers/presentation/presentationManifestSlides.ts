import type {
  AgentArtifactRepresentation,
  AgentPresentationManifest,
  AgentPresentationSlide,
} from "../../contracts";
import { isSafeArtifactImage } from "../../safeArtifactImage";

export interface RenderablePresentationSlide extends AgentPresentationSlide {
  pageRepresentationId: string;
}

export function presentationManifestSlides(
  manifest: AgentPresentationManifest,
  representations: readonly AgentArtifactRepresentation[],
): RenderablePresentationSlide[] {
  const representationMap = new Map(
    representations.map((representation) => [
      representation.representationId,
      representation,
    ]),
  );
  return manifest.slides.flatMap((slide) => {
    if (!slide.pageRepresentationId) return [];
    const page = representationMap.get(slide.pageRepresentationId);
    if (page?.status !== "ready" || !isSafeArtifactImage(page.mediaType)) {
      return [];
    }
    const thumbnail = slide.thumbnailRepresentationId
      ? representationMap.get(slide.thumbnailRepresentationId)
      : undefined;
    return [{
      ...slide,
      pageRepresentationId: slide.pageRepresentationId,
      thumbnailRepresentationId:
        thumbnail?.status === "ready" &&
        isSafeArtifactImage(thumbnail.mediaType)
          ? thumbnail.representationId
          : undefined,
    }];
  });
}

export function presentationRepresentationIds(
  slides: readonly RenderablePresentationSlide[],
): string[] {
  return [
    ...new Set(
      slides.flatMap((slide) => [
        slide.pageRepresentationId,
        ...(slide.thumbnailRepresentationId
          ? [slide.thumbnailRepresentationId]
          : []),
      ]),
    ),
  ];
}
