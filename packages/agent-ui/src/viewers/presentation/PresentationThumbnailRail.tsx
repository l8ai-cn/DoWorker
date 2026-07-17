import type { PresentationSlide } from "./presentationContracts";
import { presentationSlideLabel } from "./presentationSlideOrder";

export function PresentationThumbnailRail({
  activeSlideId,
  onSelect,
  slides,
}: {
  activeSlideId?: string;
  onSelect: (slideId: string) => void;
  slides: readonly PresentationSlide[];
}) {
  return (
    <nav
      aria-label="演示文稿页面"
      className="flex min-w-0 gap-2 overflow-x-auto border-b border-border p-2 lg:max-h-[70vh] lg:w-40 lg:flex-col lg:overflow-y-auto lg:border-b-0 lg:border-r"
      data-testid="presentation-thumbnail-rail"
    >
      {slides.map((slide, index) => {
        const label = presentationSlideLabel(slide, index);
        const active = slide.slideId === activeSlideId;
        return (
          <button
            aria-current={active ? "page" : undefined}
            aria-label={`转到第 ${index + 1} 页：${label}`}
            className={`w-28 shrink-0 rounded-md border p-1 text-left outline-none focus-visible:ring-2 focus-visible:ring-ring lg:w-full ${
              active
                ? "border-primary bg-primary/5"
                : "border-border bg-background hover:bg-muted"
            }`}
            key={slide.slideId}
            onClick={() => onSelect(slide.slideId)}
            type="button"
          >
            <div className="aspect-video overflow-hidden rounded-sm bg-muted">
              <img
                alt=""
                className="size-full object-cover"
                draggable={false}
                src={slide.thumbnailSrc ?? slide.imageSrc}
              />
            </div>
            <div className="mt-1 flex min-w-0 items-center gap-1 text-xs">
              <span className="shrink-0 text-muted-foreground">
                {index + 1}
              </span>
              <span className="truncate font-medium" title={label}>
                {label}
              </span>
            </div>
          </button>
        );
      })}
    </nav>
  );
}
