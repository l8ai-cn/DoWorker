"use client";

import React, { useCallback, useMemo, useState } from "react";
import { AgentWorkspace } from "@do-worker/agent-ui";
import { useLocale } from "next-intl";
import { cn } from "@/lib/utils";
import { useWorkspaceStore, type SplitDirection } from "@/stores/workspace";
import { usePod, usePodStore } from "@/stores/pod";
import { usePodStatus } from "@/hooks";
import { useMigratedSessionHydration } from "@/hooks/useMigratedSessionHydration";
import { AgentPanelHeader } from "./AgentPanelHeader";
import {
  PaneLoadingState,
  PaneErrorState,
} from "./PaneStateViews";
import { PodSelectorModal } from "./PodSelectorModal";
import { WorkerControlOverlay } from "@/components/mobile-worker/WorkerControlOverlay";
import { WebAcpSessionRuntime } from "./agent-ui/WebAcpSessionRuntime";

interface AgentPanelProps {
  paneId: string;
  podKey: string;
  isActive: boolean;
  onClose?: () => void;
  onMaximize?: () => void;
  onPopout?: () => void;
  showHeader?: boolean;
  allowSplit?: boolean;
  controlClientLabel?: string;
  className?: string;
}

export function AgentPanel({
  paneId,
  podKey,
  isActive,
  onClose,
  onMaximize,
  onPopout,
  showHeader = true,
  allowSplit = true,
  controlClientLabel = "desktop",
  className,
}: AgentPanelProps) {
  const locale = useLocale();
  const [isMaximized, setIsMaximized] = useState(false);
  const [pendingSplitDirection, setPendingSplitDirection] =
    useState<SplitDirection | null>(null);

  const setActivePane = useWorkspaceStore((s) => s.setActivePane);
  const splitPane = useWorkspaceStore((s) => s.splitPane);
  const panes = useWorkspaceStore((s) => s.panes);
  const pod = usePod(podKey);
  const initProgress = usePodStore((state) => state.initProgress[podKey]);

  const openPodKeys = useMemo(() => panes.map((p) => p.podKey), [panes]);
  const { podStatus, isPodReady, podError } = usePodStatus(podKey);
  const controlLease = useWorkerControlLease(podKey, controlClientLabel);

  const shouldSubscribe = isPodReady || podStatus === "running";
  const shouldShowWorkspace =
    shouldSubscribe || podStatus === "completed" || podStatus === "orphaned";
  useMigratedSessionHydration(podKey, Boolean(podKey));
  const runtime = useMemo(
    () =>
      new WebAcpSessionRuntime({
        agentLabel: pod?.agent?.name ?? "Agent",
        live: shouldSubscribe,
        paneId,
        podKey,
        title: pod?.title ?? pod?.alias ?? podKey,
      }),
    [paneId, pod?.agent?.name, pod?.alias, pod?.title, podKey, shouldSubscribe],
  );
  const handleFocus = useCallback(() => {
    setActivePane(paneId);
  }, [paneId, setActivePane]);

  const handleMaximize = useCallback(() => {
    setIsMaximized((prev) => !prev);
    onMaximize?.();
  }, [onMaximize]);

  return (
    <div
      className={cn(
        "relative flex flex-col h-full bg-background rounded-lg overflow-hidden border",
        isActive ? "border-primary" : "border-border",
        isMaximized && "fixed inset-4 z-50",
        className
      )}
      onClick={handleFocus}
    >
      {showHeader && (
        <AgentPanelHeader
          podKey={podKey}
          isMaximized={isMaximized}
          onPopout={onPopout}
          onSplitRight={allowSplit ? () => setPendingSplitDirection("horizontal") : undefined}
          onSplitDown={allowSplit ? () => setPendingSplitDirection("vertical") : undefined}
          onMaximize={handleMaximize}
          onClose={onClose}
        />
      )}

      {!shouldShowWorkspace ? (
        podError ? (
          <PaneErrorState error={podError} onClose={onClose} />
        ) : (
          <PaneLoadingState
            podStatus={podStatus}
            initProgress={initProgress}
            onClose={onClose}
          />
        )
      ) : sessionLinkLoading ? (
        <AgentSessionLinkState locale={locale} />
      ) : sessionLinkError ? (
        <PaneErrorState
          error={
            locale === "zh"
              ? `Agent 会话连接失败：${sessionLinkError}`
              : `Failed to connect to the Agent session: ${sessionLinkError}`
          }
          onClose={onClose}
        />
      ) : !runtime || !sessionId ? (
        <PaneErrorState
          error={
            locale === "zh"
              ? "该 Worker 尚未绑定真实 Agent 会话"
              : "This Worker is not linked to an Agent session"
          }
          onClose={onClose}
        />
      ) : (
        <AgentWorkspace
          className="flex-1"
          clientLabel={controlClientLabel}
          locale={locale === "zh" ? "zh-CN" : "en-US"}
          runtime={runtime}
          sessionId={runtime.sessionId}
        />
      )}
      {shouldSubscribe && (
        <WorkerControlOverlay
          podKey={podKey}
          clientLabel={controlClientLabel}
          preserveHeader={showHeader}
        />
      )}

      {pendingSplitDirection && (
        <PodSelectorModal
          openPodKeys={openPodKeys}
          onSelect={(selectedPodKey) => {
            splitPane(paneId, pendingSplitDirection, selectedPodKey);
            setPendingSplitDirection(null);
          }}
          onClose={() => setPendingSplitDirection(null)}
        />
      )}
    </div>
  );
}

export default AgentPanel;
