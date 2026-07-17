import type { PresentationSlide } from "./presentationContracts";

export function orderPresentationSlides(
  slides: readonly PresentationSlide[],
): PresentationSlide[] {
  return [...slides].sort(
    (left, right) =>
      left.position - right.position ||
      left.slideId.localeCompare(right.slideId),
  );
}

export function presentationSlideLabel(
  slide: PresentationSlide,
  index: number,
): string {
  return slide.title?.trim() || `第 ${index + 1} 页`;
}
