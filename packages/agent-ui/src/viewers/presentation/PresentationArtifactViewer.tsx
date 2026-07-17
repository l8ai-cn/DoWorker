import { StickyNote } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { PresentationActionToolbar } from "./PresentationActionToolbar";
import { createPresentationAction } from "./presentationAction";
import {
  PRESENTATION_GRANTS,
  type PresentationArtifactViewerProps,
  type PresentationGrant,
} from "./presentationContracts";
import { orderPresentationSlides } from "./presentationSlideOrder";
import { PresentationStage } from "./PresentationStage";
import { PresentationThumbnailRail } from "./PresentationThumbnailRail";

export {
  PRESENTATION_GRANTS,
  type PresentationArtifactAction,
  type PresentationArtifactViewerProps,
  type PresentationGrant,
  type PresentationSlide,
  type PresentationVersion,
} from "./presentationContracts";

export function PresentationArtifactViewer({
  actionSchemaVersion,
  artifactId,
  baseRevision,
  grants,
  initialSlideId,
  onAction,
  onRequestFullscreen,
  onSelectVersion,
  representationId,
  selectedVersionId,
  slides,
  versions,
}: PresentationArtifactViewerProps) {
  const orderedSlides = useMemo(
    () => orderPresentationSlides(slides),
    [slides],
  );
  const [currentSlideId, setCurrentSlideId] = useState(
    () =>
      orderedSlides.find((slide) => slide.slideId === initialSlideId)
        ?.slideId ?? orderedSlides[0]?.slideId,
  );

  useEffect(() => {
    if (!orderedSlides.some((slide) => slide.slideId === currentSlideId)) {
      setCurrentSlideId(
        orderedSlides.find((slide) => slide.slideId === initialSlideId)
          ?.slideId ?? orderedSlides[0]?.slideId,
      );
    }
  }, [currentSlideId, initialSlideId, orderedSlides]);

  const currentIndex = orderedSlides.findIndex(
    (slide) => slide.slideId === currentSlideId,
  );
  const currentSlide = orderedSlides[currentIndex];
  const emitAction = (
    actionType: PresentationGrant,
    payload:
      | { slideId: string }
      | { slideId: string; targetIndex: number }
      | { format: "pptx"; slideId: string },
  ) => {
    if (!currentSlide) return;
    onAction(
      createPresentationAction(
        {
          actionSchemaVersion,
          artifactId,
          baseRevision,
          representationId:
            actionType === PRESENTATION_GRANTS.exportPresentation
              ? representationId
              : currentSlide.representationId ?? representationId,
        },
        actionType,
        payload,
      ),
    );
  };

  return (
    <section className="overflow-hidden rounded-md border border-border bg-background text-foreground">
      <header className="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-border px-3 py-2.5">
        <div className="min-w-0">
          <div className="text-sm font-semibold">演示文稿</div>
          <div className="text-xs text-muted-foreground">
            修订 {baseRevision.toString()}
          </div>
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-muted-foreground">
            <span>版本</span>
            <select
              aria-label="选择演示文稿版本"
              className="h-9 max-w-44 rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!onSelectVersion || versions.length < 2}
              onChange={(event) =>
                onSelectVersion?.(event.currentTarget.value)
              }
              value={selectedVersionId}
            >
              {versions.map((version) => (
                <option key={version.id} value={version.id}>
                  {version.label}
                </option>
              ))}
            </select>
          </label>
          <PresentationActionToolbar
            currentIndex={currentIndex}
            grants={grants}
            onExport={() =>
              currentSlide &&
              emitAction(PRESENTATION_GRANTS.exportPresentation, {
                format: "pptx",
                slideId: currentSlide.slideId,
              })
            }
            onMove={(targetIndex) =>
              currentSlide &&
              emitAction(PRESENTATION_GRANTS.reorderSlide, {
                slideId: currentSlide.slideId,
                targetIndex,
              })
            }
            onRegenerate={() =>
              currentSlide &&
              emitAction(PRESENTATION_GRANTS.regenerateSlide, {
                slideId: currentSlide.slideId,
              })
            }
            onReplace={() =>
              currentSlide &&
              emitAction(PRESENTATION_GRANTS.replaceSlide, {
                slideId: currentSlide.slideId,
              })
            }
            slideCount={orderedSlides.length}
          />
        </div>
      </header>
      <div className="lg:grid lg:grid-cols-[10rem_minmax(0,1fr)_16rem]">
        <PresentationThumbnailRail
          activeSlideId={currentSlide?.slideId}
          onSelect={setCurrentSlideId}
          slides={orderedSlides}
        />
        <PresentationStage
          currentIndex={currentIndex}
          currentSlide={currentSlide}
          onRequestFullscreen={onRequestFullscreen}
          slideCount={orderedSlides.length}
        />
        <aside className="min-w-0 border-t border-border p-3 lg:border-l lg:border-t-0">
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            <StickyNote aria-hidden="true" className="size-4" />
            讲者备注
          </h2>
          <div className="mt-3 whitespace-pre-wrap text-sm leading-6 text-muted-foreground">
            {currentSlide?.notes?.trim() || "暂无讲者备注"}
          </div>
        </aside>
      </div>
    </section>
  );
}
