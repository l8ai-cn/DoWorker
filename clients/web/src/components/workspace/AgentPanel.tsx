"use client";

import React, { useCallback, useMemo, useState } from "react";
import {
  AgentWorkspace,
  createBuiltinContentRenderers,
} from "@do-worker/agent-ui";
import { useLocale } from "next-intl";
import { cn } from "@/lib/utils";
import { useWorkspaceStore, type SplitDirection } from "@/stores/workspace";
import { usePod, usePodStore } from "@/stores/pod";
import { usePodStatus } from "@/hooks";
import { useAcpRelay } from "@/hooks/useAcpRelay";
import { useAgentSessionLink } from "@/hooks/useAgentSessionLink";
import { getAgentWorkbenchState } from "@/lib/wasm-core";
import { AgentPanelHeader } from "./AgentPanelHeader";
import { AgentSessionLinkState } from "./AgentSessionLinkState";
import {
  PaneLoadingState,
  PaneErrorState,
  PaneReconnectingState,
} from "./PaneStateViews";
import { PodSelectorModal } from "./PodSelectorModal";
import { WorkerControlOverlay } from "@/components/mobile-worker/WorkerControlOverlay";
import { useWorkerControlLease } from "@/hooks/useWorkerControlLease";
import { WebAgentWorkbenchRuntime } from "./agent-ui/WebAgentWorkbenchRuntime";
import {
  createWebAgentWorkbenchArtifactLoader,
} from "./agent-ui/webAgentWorkbenchArtifactLoader";

const AGENT_CONTENT_RENDERERS = createBuiltinContentRenderers();

interface AgentPanelProps {
  paneId: string;
  podKey: string;
  isActive: boolean;
  onClose?: () => void;
  onMaximize?: () => void;
  onPopout?: () => void;
  showHeader?: boolean;
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

  const liveSession = isPodReady || podStatus === "running";
  const canReadSession = liveSession || podStatus === "completed";
  useAcpRelay(podKey, paneId, liveSession);
  const {
    error: sessionLinkError,
    loading: sessionLinkLoading,
    sessionId,
  } = useAgentSessionLink(podKey, canReadSession);
  const runtime = useMemo(
    () => {
      if (!sessionId) return null;
      const workbenchState = getAgentWorkbenchState();
      return new WebAgentWorkbenchRuntime({
        agentLabel: pod?.agent?.name ?? "Agent",
        interactionMode: pod?.interaction_mode ?? "acp",
        loadArtifact: createWebAgentWorkbenchArtifactLoader(workbenchState),
        live: liveSession,
        sessionId,
        title: pod?.title ?? pod?.alias ?? podKey,
      });
    },
    [
      pod?.agent?.name,
      pod?.alias,
      pod?.interaction_mode,
      pod?.title,
      podKey,
      liveSession,
      sessionId,
    ],
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
        !showHeader && controlLease.status !== "granted" && "max-sm:pb-20",
        className
      )}
      onClick={handleFocus}
    >
      {showHeader && (
        <AgentPanelHeader
          podKey={podKey}
          isMaximized={isMaximized}
          onPopout={onPopout}
          onSplitRight={() => setPendingSplitDirection("horizontal")}
          onSplitDown={() => setPendingSplitDirection("vertical")}
          onMaximize={handleMaximize}
          onClose={onClose}
        />
      )}

      {!canReadSession ? (
        podError ? (
          <PaneErrorState
            error={
              locale === "zh"
                ? "Worker 启动失败，请稍后重试"
                : "The Worker failed to start. Please try again."
            }
            onClose={onClose}
          />
        ) : podStatus === "orphaned" ? (
          <PaneReconnectingState onClose={onClose} />
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
              ? "Worker 会话连接失败，请稍后重试"
              : "Failed to connect to the Worker session. Please try again."
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
          contentRenderers={AGENT_CONTENT_RENDERERS}
          locale={locale === "zh" ? "zh-CN" : "en-US"}
          presentation="user"
          readOnly={!liveSession || controlLease.status !== "granted"}
          runtime={runtime}
          sessionId={sessionId}
        />
      )}
      {liveSession && (
        <WorkerControlOverlay
          blocking={false}
          lease={controlLease}
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
