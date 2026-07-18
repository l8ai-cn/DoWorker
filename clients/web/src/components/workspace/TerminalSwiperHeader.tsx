"use client";

import { ChevronLeft, ChevronRight, Scaling } from "lucide-react";
import { Button } from "@/components/ui/button";
import { usePodTitle } from "@/hooks/usePodTitle";
import { POD_MODE_PTY } from "@/lib/pod-modes";
import { usePod } from "@/stores/pod";

interface TerminalSwiperHeaderProps {
  activeIndex: number;
  onNext: () => void;
  onPrev: () => void;
  onSyncSize: () => void;
  paneCount: number;
  podKey?: string;
}

export function TerminalSwiperHeader({
  activeIndex,
  onNext,
  onPrev,
  onSyncSize,
  paneCount,
  podKey,
}: TerminalSwiperHeaderProps) {
  const pod = usePod(podKey);
  const isTerminal = pod?.interaction_mode === POD_MODE_PTY;

  return (
    <div className="flex h-10 items-center justify-between border-b border-terminal-border bg-terminal-bg-secondary px-3">
      <Button
        className="h-7 w-7 p-0 text-terminal-text-muted"
        disabled={activeIndex === 0}
        onClick={onPrev}
        size="sm"
        variant="ghost"
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>

      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-terminal-text">
          {podKey ? <SwiperPaneTitle podKey={podKey} /> : "Worker"}
        </span>
        <span className="text-xs text-terminal-text-muted">
          {activeIndex + 1} / {paneCount}
        </span>
      </div>

      <div className="flex items-center gap-1">
        <Button
          aria-hidden={!isTerminal}
          className={`h-7 w-7 p-0 text-terminal-text-muted ${
            isTerminal ? "" : "invisible"
          }`}
          disabled={!isTerminal}
          onClick={onSyncSize}
          size="sm"
          title="Sync terminal size"
          variant="ghost"
        >
          <Scaling className="h-4 w-4" />
        </Button>
        <Button
          className="h-7 w-7 p-0 text-terminal-text-muted"
          disabled={activeIndex === paneCount - 1}
          onClick={onNext}
          size="sm"
          variant="ghost"
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function SwiperPaneTitle({ podKey }: { podKey: string }) {
  const title = usePodTitle(podKey, "Worker");
  return <>{title}</>;
}
