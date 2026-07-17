import {
  ArrowDown,
  ArrowUp,
  Download,
  RefreshCw,
  Replace,
  type LucideIcon,
} from "lucide-react";

import {
  PRESENTATION_GRANTS,
  type PresentationGrant,
} from "./presentationContracts";

export function PresentationActionToolbar({
  currentIndex,
  grants,
  onExport,
  onMove,
  onRegenerate,
  onReplace,
  slideCount,
}: {
  currentIndex: number;
  grants: readonly PresentationGrant[];
  onExport: () => void;
  onMove: (targetIndex: number) => void;
  onRegenerate: () => void;
  onReplace: () => void;
  slideCount: number;
}) {
  const hasSlide = currentIndex >= 0;
  const granted = new Set(grants);
  return (
    <div
      aria-label="演示文稿操作"
      className="flex min-w-0 flex-wrap items-center gap-1"
      role="toolbar"
    >
      <ActionButton
        disabled={
          !hasSlide || !granted.has(PRESENTATION_GRANTS.regenerateSlide)
        }
        icon={RefreshCw}
        label="重新生成当前页"
        onClick={onRegenerate}
      />
      <ActionButton
        disabled={!hasSlide || !granted.has(PRESENTATION_GRANTS.replaceSlide)}
        icon={Replace}
        label="替换当前页"
        onClick={onReplace}
      />
      <ActionButton
        disabled={
          currentIndex <= 0 || !granted.has(PRESENTATION_GRANTS.reorderSlide)
        }
        icon={ArrowUp}
        label="上移当前页"
        onClick={() => onMove(currentIndex - 1)}
      />
      <ActionButton
        disabled={
          currentIndex < 0 ||
          currentIndex >= slideCount - 1 ||
          !granted.has(PRESENTATION_GRANTS.reorderSlide)
        }
        icon={ArrowDown}
        label="下移当前页"
        onClick={() => onMove(currentIndex + 1)}
      />
      <ActionButton
        disabled={
          !hasSlide || !granted.has(PRESENTATION_GRANTS.exportPresentation)
        }
        icon={Download}
        label="导出演示文稿"
        onClick={onExport}
      />
    </div>
  );
}

function ActionButton({
  disabled,
  icon: Icon,
  label,
  onClick,
}: {
  disabled: boolean;
  icon: LucideIcon;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={label}
      className="inline-flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-40"
      disabled={disabled}
      onClick={onClick}
      title={label}
      type="button"
    >
      <Icon aria-hidden="true" className="size-4" />
    </button>
  );
}
