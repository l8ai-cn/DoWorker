import { useRef, useState } from "react";

import type { PresentationSlide } from "./presentationContracts";
import { presentationSlideLabel } from "./presentationSlideOrder";
import { PresentationViewControls } from "./PresentationViewControls";

export function PresentationStage({
  currentIndex,
  currentSlide,
  onRequestFullscreen,
  slideCount,
}: {
  currentIndex: number;
  currentSlide?: PresentationSlide;
  onRequestFullscreen?: () => void | Promise<void>;
  slideCount: number;
}) {
  const [zoomPercent, setZoomPercent] = useState(100);
  const [fitToWindow, setFitToWindow] = useState(true);
  const stageRef = useRef<HTMLDivElement>(null);
  const fullscreenSupported =
    Boolean(onRequestFullscreen) ||
    (typeof HTMLElement !== "undefined" &&
      typeof HTMLElement.prototype.requestFullscreen === "function");

  return (
    <main className="min-w-0 p-3">
      <div className="mb-2 flex min-w-0 flex-wrap items-center justify-between gap-2">
        <span className="text-xs font-medium tabular-nums">
          {currentSlide
            ? `第 ${currentIndex + 1} / ${slideCount} 页`
            : "暂无页面"}
        </span>
        <PresentationViewControls
          canFullscreen={fullscreenSupported}
          disabled={!currentSlide}
          fitToWindow={fitToWindow}
          onFitToWindow={() => {
            setFitToWindow(true);
            setZoomPercent(100);
          }}
          onRequestFullscreen={() => {
            if (onRequestFullscreen) {
              void onRequestFullscreen();
              return;
            }
            void stageRef.current?.requestFullscreen();
          }}
          onZoomChange={(value) => {
            setFitToWindow(false);
            setZoomPercent(value);
          }}
          zoomPercent={zoomPercent}
        />
      </div>
      <div
        className="flex aspect-video min-h-52 min-w-0 w-full items-center justify-center overflow-auto rounded-md border border-border bg-muted/30"
        ref={stageRef}
      >
        {currentSlide ? (
          <img
            alt={`第 ${currentIndex + 1} 页：${presentationSlideLabel(
              currentSlide,
              currentIndex,
            )}`}
            className="max-h-full max-w-full object-contain transition-transform motion-reduce:transition-none"
            data-testid="presentation-slide-image"
            draggable={false}
            src={currentSlide.imageSrc}
            style={{
              transform: `scale(${zoomPercent / 100})`,
              transformOrigin: "center",
            }}
          />
        ) : (
          <span className="text-sm text-muted-foreground">暂无可预览页面</span>
        )}
      </div>
    </main>
  );
}
