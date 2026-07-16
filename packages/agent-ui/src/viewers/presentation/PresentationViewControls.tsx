import { Maximize2, Scan, ZoomIn } from "lucide-react";

export function PresentationViewControls({
  canFullscreen,
  disabled,
  fitToWindow,
  onFitToWindow,
  onRequestFullscreen,
  onZoomChange,
  zoomPercent,
}: {
  canFullscreen: boolean;
  disabled: boolean;
  fitToWindow: boolean;
  onFitToWindow: () => void;
  onRequestFullscreen: () => void;
  onZoomChange: (zoomPercent: number) => void;
  zoomPercent: number;
}) {
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <button
        aria-pressed={fitToWindow}
        className="inline-flex h-9 items-center gap-1.5 rounded-md border border-border px-2.5 text-xs font-medium outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled}
        onClick={onFitToWindow}
        type="button"
      >
        <Scan aria-hidden="true" className="size-3.5" />
        适应窗口
      </button>
      <label className="flex h-9 min-w-44 items-center gap-2 text-xs text-muted-foreground">
        <ZoomIn aria-hidden="true" className="size-3.5 shrink-0" />
        <input
          aria-label="缩放比例"
          className="min-w-24 flex-1 accent-primary"
          disabled={disabled}
          max={200}
          min={50}
          onChange={(event) => onZoomChange(Number(event.currentTarget.value))}
          step={10}
          type="range"
          value={zoomPercent}
        />
        <span className="w-10 text-right tabular-nums">{zoomPercent}%</span>
      </label>
      <button
        className="inline-flex size-9 items-center justify-center rounded-md border border-border outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled || !canFullscreen}
        onClick={onRequestFullscreen}
        title={canFullscreen ? "全屏查看" : "当前环境不支持全屏"}
        type="button"
      >
        <Maximize2 aria-hidden="true" className="size-4" />
        <span className="sr-only">全屏查看</span>
      </button>
    </div>
  );
}
