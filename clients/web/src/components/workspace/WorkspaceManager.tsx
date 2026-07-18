"use client";

import React, { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useBreakpoint } from "@/components/layout/useBreakpoint";
import { useWorkspaceStore } from "@/stores/workspace";
import { usePod } from "@/stores/pod";
import { POD_MODE_PTY } from "@/lib/pod-modes";
import { TerminalGrid } from "./TerminalGrid";
import { TerminalSwiper } from "./TerminalSwiper";
import { TerminalToolbar } from "./TerminalToolbar";
import { PodSelectorModal } from "./PodSelectorModal";

interface WorkspaceManagerProps {
  className?: string;
}

export function WorkspaceManager({ className }: WorkspaceManagerProps) {
  const { isMobile } = useBreakpoint();
  const panes = useWorkspaceStore((s) => s.panes);
  const mobileActiveIndex = useWorkspaceStore((s) => s.mobileActiveIndex);
  const addPane = useWorkspaceStore((s) => s.addPane);
  const _hasHydrated = useWorkspaceStore((s) => s._hasHydrated);
  const [showPodSelector, setShowPodSelector] = useState(false);

  const openPodKeys = useMemo(() => panes.map((p) => p.podKey), [panes]);
  const mobilePod = usePod(panes[mobileActiveIndex]?.podKey);

  const handleAddNew = () => {
    setShowPodSelector(true);
  };

  const handleSelectPod = (podKey: string) => {
    addPane(podKey);
    setShowPodSelector(false);
  };

  const handlePopout = (paneId: string) => {
    const pane = panes.find((p) => p.id === paneId);
    if (!pane) return;

    const popoutUrl = `/popout/terminal/${pane.podKey}`;
    const popoutWindow = window.open(
      popoutUrl,
      `terminal-${pane.podKey}`,
      "width=800,height=600,menubar=no,toolbar=no,location=no,status=no"
    );

    if (popoutWindow) {
    }
  };

  if (!_hasHydrated) {
    return (
      <div className="h-full bg-terminal-bg">
        <CenteredSpinner />
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col h-full bg-terminal-bg", className)}>
      {!isMobile && (
        <TerminalGrid
          onPopout={handlePopout}
          onAddNew={handleAddNew}
          className="flex-1"
        />
      )}

      {isMobile && (
        <>
          <TerminalSwiper onAddNew={handleAddNew} className="flex-1" />
          {mobilePod?.interaction_mode === POD_MODE_PTY && <TerminalToolbar />}
        </>
      )}

      {showPodSelector && (
        <PodSelectorModal
          openPodKeys={openPodKeys}
          onSelect={handleSelectPod}
          onClose={() => setShowPodSelector(false)}
        />
      )}
    </div>
  );
}

export default WorkspaceManager;
