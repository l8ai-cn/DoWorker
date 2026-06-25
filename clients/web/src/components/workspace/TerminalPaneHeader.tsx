"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { AgentStatusBadge } from "@/components/shared/AgentStatusBadge";
import { usePod } from "@/stores/pod";
import { usePodTitle } from "@/hooks/usePodTitle";
import { getShortPodKey } from "@/lib/pod-display-name";
import type { ConnectionStatus } from "@/stores/relayConnection";
import {
  X,
  Maximize2,
  Minimize2,
  ExternalLink,
  Circle,
  Scaling,
  Bot,
  PanelRight,
  PanelBottom,
} from "lucide-react";

interface TerminalPaneHeaderProps {
  podKey: string;
  connectionStatus: ConnectionStatus;
  isRunnerDisconnected?: boolean;
  isMaximized: boolean;
  isPodReady: boolean;
  hasAutopilot: boolean;
  onSyncSize: () => void;
  onStartAutopilot: () => void;
  onPopout?: () => void;
  onSplitRight?: () => void;
  onSplitDown?: () => void;
  onMaximize: () => void;
  onClose?: () => void;
}

export function TerminalPaneHeader({
  podKey,
  connectionStatus,
  isRunnerDisconnected,
  isMaximized,
  isPodReady,
  hasAutopilot,
  onSyncSize,
  onStartAutopilot,
  onPopout,
  onSplitRight,
  onSplitDown,
  onMaximize,
  onClose,
}: TerminalPaneHeaderProps) {
  const title = usePodTitle(podKey);

  const statusColor = (() => {
    if (isRunnerDisconnected) return "text-danger";
    switch (connectionStatus) {
      case "connected": return "text-success";
      case "connecting": return "text-warning animate-pulse";
      case "error": return "text-danger";
      default: return "text-muted-foreground";
    }
  })();

  return (
    <div className="h-8 flex items-center justify-between px-2 bg-terminal-bg-secondary shadow-[inset_0_-1px_0_color-mix(in_srgb,var(--terminal-border)_75%,transparent)]">
      <div className="flex items-center gap-2 min-w-0">
        <Circle className={cn("w-2 h-2 flex-shrink-0", statusColor)} />
        <span className="text-xs text-terminal-text truncate">{title}</span>
        <code className="text-[10px] text-terminal-text-muted truncate">
          {getShortPodKey(podKey)}
        </code>
        <TerminalAgentStatus podKey={podKey} />
      </div>
      <HeaderActions
        isPodReady={isPodReady}
        isMaximized={isMaximized}
        hasAutopilot={hasAutopilot}
        onSyncSize={onSyncSize}
        onStartAutopilot={onStartAutopilot}
        onSplitRight={onSplitRight}
        onSplitDown={onSplitDown}
        onPopout={onPopout}
        onMaximize={onMaximize}
        onClose={onClose}
      />
    </div>
  );
}

function HeaderActions({
  isPodReady,
  isMaximized,
  hasAutopilot,
  onSyncSize,
  onStartAutopilot,
  onSplitRight,
  onSplitDown,
  onPopout,
  onMaximize,
  onClose,
}: {
  isPodReady: boolean;
  isMaximized: boolean;
  hasAutopilot: boolean;
  onSyncSize: () => void;
  onStartAutopilot: () => void;
  onSplitRight?: () => void;
  onSplitDown?: () => void;
  onPopout?: () => void;
  onMaximize: () => void;
  onClose?: () => void;
}) {
  return (
    <div className="flex items-center gap-1 flex-shrink-0">
      <IconButton onClick={onSyncSize} title="Sync terminal size" icon={<Scaling className="w-3 h-3" />} />
      {!hasAutopilot && isPodReady && (
        <IconButton onClick={onStartAutopilot} title="Start Autopilot Mode" icon={<Bot className="w-3 h-3" />} hoverClass="hover:text-primary" />
      )}
      {onSplitRight && <IconButton onClick={onSplitRight} title="Split Right" icon={<PanelRight className="w-3 h-3" />} />}
      {onSplitDown && <IconButton onClick={onSplitDown} title="Split Down" icon={<PanelBottom className="w-3 h-3" />} />}
      {onPopout && <IconButton onClick={onPopout} title="Popout" icon={<ExternalLink className="w-3 h-3" />} />}
      <IconButton
        onClick={onMaximize}
        title={isMaximized ? "Restore" : "Maximize"}
        icon={isMaximized ? <Minimize2 className="w-3 h-3" /> : <Maximize2 className="w-3 h-3" />}
      />
      {onClose && (
        <IconButton onClick={onClose} title="Close" icon={<X className="w-3 h-3" />} hoverClass="hover:text-danger" />
      )}
    </div>
  );
}

function IconButton({ onClick, title, icon, hoverClass }: {
  onClick: () => void;
  title: string;
  icon: React.ReactNode;
  hoverClass?: string;
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      className={cn("h-5 w-5 p-0 hover:bg-terminal-bg-active text-terminal-text", hoverClass)}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
      title={title}
    >
      {icon}
    </Button>
  );
}

function TerminalAgentStatus({ podKey }: { podKey: string }) {
  const pod = usePod(podKey);
  if (!pod) return null;
  return (
    <AgentStatusBadge
      agentStatus={pod.agent_status ?? ''}
      podStatus={pod.status}
      variant="inline"
    />
  );
}

export default TerminalPaneHeader;
