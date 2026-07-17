import {
  Columns2,
  Image as ImageIcon,
  Images,
  SlidersHorizontal,
  type LucideIcon,
} from "lucide-react";
import { useState } from "react";

export type ImageComparisonMode =
  | "source"
  | "result"
  | "side-by-side"
  | "slider";

export interface ComparisonImage {
  alt: string;
  src: string;
}

export interface ImageComparisonViewerProps {
  defaultMode?: ImageComparisonMode;
  result: ComparisonImage;
  source: ComparisonImage;
}

const MODES: Array<{
  icon: LucideIcon;
  label: string;
  value: ImageComparisonMode;
}> = [
  { icon: ImageIcon, label: "查看源图", value: "source" },
  { icon: Images, label: "查看结果图", value: "result" },
  { icon: Columns2, label: "并排比较", value: "side-by-side" },
  { icon: SlidersHorizontal, label: "滑块比较", value: "slider" },
];

export function ImageComparisonViewer({
  defaultMode = "source",
  result,
  source,
}: ImageComparisonViewerProps) {
  const [mode, setMode] = useState<ImageComparisonMode>(defaultMode);
  const [sliderPosition, setSliderPosition] = useState(50);

  return (
    <section className="space-y-3">
      <div
        aria-label="图片比较模式"
        className="flex flex-wrap items-center gap-1 rounded-md border border-border bg-muted/30 p-1"
        role="group"
      >
        {MODES.map((option) => (
          <ModeButton
            active={mode === option.value}
            icon={option.icon}
            key={option.value}
            label={option.label}
            onClick={() => setMode(option.value)}
          />
        ))}
      </div>
      {mode === "source" && <ImageFrame image={source} />}
      {mode === "result" && <ImageFrame image={result} />}
      {mode === "side-by-side" && (
        <div className="grid min-w-0 gap-3 sm:grid-cols-2">
          <ImageFrame image={source} />
          <ImageFrame image={result} />
        </div>
      )}
      {mode === "slider" && (
        <div className="space-y-2">
          <div
            className="relative aspect-video w-full overflow-hidden rounded-md border border-border bg-muted"
            data-testid="image-comparison-slider-viewport"
          >
            <ComparisonImageElement image={source} />
            <div
              className="absolute inset-0 overflow-hidden"
              data-testid="image-comparison-result-layer"
              style={{
                clipPath: `inset(0 ${100 - sliderPosition}% 0 0)`,
              }}
            >
              <ComparisonImageElement image={result} />
            </div>
            <div
              aria-hidden="true"
              className="pointer-events-none absolute inset-y-0 w-px bg-primary"
              style={{ left: `${sliderPosition}%` }}
            />
          </div>
          <input
            aria-label="比较位置"
            className="h-11 w-full accent-primary"
            max={100}
            min={0}
            onChange={(event) => setSliderPosition(Number(event.target.value))}
            type="range"
            value={sliderPosition}
          />
        </div>
      )}
    </section>
  );
}

function ModeButton({
  active,
  icon: Icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: LucideIcon;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-pressed={active}
      className={`inline-flex h-9 items-center gap-1.5 rounded-md px-3 text-xs font-medium outline-none focus-visible:ring-2 focus-visible:ring-ring ${
        active
          ? "bg-background text-foreground shadow-sm"
          : "text-muted-foreground hover:bg-muted hover:text-foreground"
      }`}
      onClick={onClick}
      type="button"
    >
      <Icon className="size-3.5 shrink-0" />
      {label}
    </button>
  );
}

function ImageFrame({ image }: { image: ComparisonImage }) {
  return (
    <div className="flex min-h-56 min-w-0 items-center justify-center overflow-hidden rounded-md border border-border bg-muted">
      <ComparisonImageElement image={image} />
    </div>
  );
}

function ComparisonImageElement({ image }: { image: ComparisonImage }) {
  return (
    <img
      alt={image.alt}
      className="size-full object-contain"
      draggable={false}
      src={image.src}
    />
  );
}
