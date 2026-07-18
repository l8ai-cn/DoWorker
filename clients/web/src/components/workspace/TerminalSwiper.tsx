"use client";

import React, { useEffect, useState } from "react";
import { useDrag } from "@use-gesture/react";
import { cn } from "@/lib/utils";
import { useWorkspaceStore } from "@/stores/workspace";
import { useTerminalInput } from "@/hooks/useTerminalInput";
import { Terminal as TerminalIcon, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { MobileWorkspacePane } from "./MobileWorkspacePane";
import { TerminalSwiperHeader } from "./TerminalSwiperHeader";

interface TerminalSwiperProps {
  onAddNew?: () => void;
  className?: string;
}

export function TerminalSwiper({ onAddNew, className }: TerminalSwiperProps) {
  const panes = useWorkspaceStore((s) => s.panes);
  const mobileActiveIndex = useWorkspaceStore((s) => s.mobileActiveIndex);
  const setMobileActiveIndex = useWorkspaceStore((s) => s.setMobileActiveIndex);
  const removePane = useWorkspaceStore((s) => s.removePane);
  const { syncSize } = useTerminalInput();

  const [translateX, setTranslateX] = useState(0);
  const [isDragging, setIsDragging] = useState(false);

  const bind = useDrag(
    ({ movement: [mx, my], direction: [dx], velocity: [vx], last, cancel, event }) => {
      if (panes.length <= 1) return;

      const target = event?.target as HTMLElement;
      if (target?.closest('.xterm-helper-textarea') || target?.closest('.xterm-screen')) {
        cancel();
        setTranslateX(0);
        setIsDragging(false);
        return;
      }

      if (!last && Math.abs(my) > Math.abs(mx) * 1.2) {
        cancel();
        setTranslateX(0);
        setIsDragging(false);
        return;
      }

      setIsDragging(!last);

      if (last) {
        const threshold = 50;
        const velocityThreshold = 0.5;

        let newIndex = mobileActiveIndex;

        if (mx < -threshold || (vx > velocityThreshold && dx < 0)) {
          newIndex = Math.min(mobileActiveIndex + 1, panes.length - 1);
        } else if (mx > threshold || (vx > velocityThreshold && dx > 0)) {
          newIndex = Math.max(mobileActiveIndex - 1, 0);
        }

        setMobileActiveIndex(newIndex);
        setTranslateX(0);
      } else {
        const maxDrag = 100;
        setTranslateX(Math.max(-maxDrag, Math.min(maxDrag, mx)));
      }
    },
    {
      axis: "lock",
      filterTaps: true,
      rubberband: true,
      threshold: 10,
    }
  );

  const goToPrev = () => {
    if (mobileActiveIndex > 0) {
      setMobileActiveIndex(mobileActiveIndex - 1);
    }
  };

  const goToNext = () => {
    if (mobileActiveIndex < panes.length - 1) {
      setMobileActiveIndex(mobileActiveIndex + 1);
    }
  };

  useEffect(() => {
    if (panes.length === 0) return;

    if (mobileActiveIndex >= panes.length) {
      setMobileActiveIndex(panes.length - 1);
    } else if (mobileActiveIndex < 0) {
      setMobileActiveIndex(0);
    }
  }, [panes.length, mobileActiveIndex, setMobileActiveIndex]);

  const currentPane = panes[mobileActiveIndex];

  if (panes.length === 0) {
    return (
      <div className={cn("flex-1 flex items-center justify-center bg-terminal-bg", className)}>
        <div className="text-center p-6">
          <TerminalIcon className="w-16 h-16 mx-auto mb-4 text-terminal-border" />
          <h3 className="text-lg font-medium text-terminal-text mb-2">No terminals</h3>
          <p className="text-sm text-terminal-text-muted mb-4">
            Open a pod to start a terminal session
          </p>
          {onAddNew && (
            <Button onClick={onAddNew}>
              <Plus className="w-4 h-4 mr-2" />
              Open Terminal
            </Button>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col h-full", className)}>
      <TerminalSwiperHeader
        activeIndex={mobileActiveIndex}
        podKey={currentPane?.podKey}
        paneCount={panes.length}
        onPrev={goToPrev}
        onNext={goToNext}
        onSyncSize={syncSize}
      />

      {panes.length > 1 && (
        <div
          {...bind()}
          className="flex items-center justify-center gap-1.5 py-2 bg-terminal-bg-secondary"
          style={{ touchAction: "pan-y" }}
        >
          {panes.map((pane, index) => (
            <button
              key={pane.id}
              className={cn(
                "w-1.5 h-1.5 rounded-full transition-colors",
                index === mobileActiveIndex
                  ? "bg-primary"
                  : "bg-terminal-border hover:bg-terminal-bg-active"
              )}
              onClick={() => setMobileActiveIndex(index)}
            />
          ))}
        </div>
      )}

      <div
        className="flex-1 overflow-hidden"
        style={{
          transform: isDragging ? `translateX(${translateX}px)` : "none",
          transition: isDragging ? "none" : "transform 0.2s ease-out",
        }}
      >
        {currentPane && (
          <MobileWorkspacePane
            paneId={currentPane.id}
            podKey={currentPane.podKey}
            onClose={() => removePane(currentPane.id)}
          />
        )}
      </div>
    </div>
  );
}

export default TerminalSwiper;
