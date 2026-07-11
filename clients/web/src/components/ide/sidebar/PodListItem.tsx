"use client";

import { cn } from "@/lib/utils";
import { getPodDisplayName } from "@/lib/pod-display-name";
import { Pod } from "@/stores/pod";
import { AgentStatusBadge } from "@/components/shared/AgentStatusBadge";
import { Button } from "@/components/ui/button";
import { useTranslations } from "next-intl";
import {
  Square,
  Terminal,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  RefreshCw,
  Smartphone,
} from "lucide-react";
import { SidebarPodContextMenu } from "./SidebarPodContextMenu";
import { SidebarPodActionsMenu } from "./SidebarPodActionsMenu";

const statusColors: Record<string, { bg: string; text: string; dot: string }> = {
  initializing: { bg: "bg-warning-bg", text: "text-warning", dot: "bg-warning" },
  running: { bg: "bg-info-bg", text: "text-info", dot: "bg-info" },
  paused: { bg: "bg-accent", text: "text-primary", dot: "bg-primary" },
  disconnected: { bg: "bg-muted", text: "text-muted-foreground", dot: "bg-muted-foreground" },
  orphaned: { bg: "bg-warning-bg", text: "text-warning", dot: "bg-warning" },
  completed: { bg: "bg-success-bg", text: "text-success", dot: "bg-success" },
  terminated: { bg: "bg-muted", text: "text-muted-foreground", dot: "bg-muted-foreground" },
  error: { bg: "bg-danger-bg", text: "text-danger", dot: "bg-danger" },
  failed: { bg: "bg-danger-bg", text: "text-danger", dot: "bg-danger" },
};

function getStatusIcon(status: string) {
  switch (status) {
    case "initializing":
      return <Clock className="w-3 h-3" />;
    case "running":
      return <Loader2 className="w-3 h-3 animate-spin" />;
    case "orphaned":
      return <RefreshCw className="w-3 h-3 animate-spin" />;
    case "paused":
      return <Square className="w-3 h-3" />;
    case "terminated":
      return <CheckCircle className="w-3 h-3" />;
    case "failed":
      return <XCircle className="w-3 h-3" />;
    default:
      return <Square className="w-3 h-3" />;
  }
}

interface PodListItemProps {
  pod: Pod;
  isOpen: boolean;
  onClick: () => void;
  onTerminate: () => void;
  onDelete: () => void;
  onWake: () => void;
  onRename: () => void;
  onShare: () => void;
  onOpenMobile: () => void;
  onPublishExpert?: () => void;
  onTogglePerpetual: (perpetual: boolean) => void;
}

export function PodListItem({ pod, isOpen, onClick, onTerminate, onDelete, onWake, onRename, onShare, onOpenMobile, onPublishExpert, onTogglePerpetual }: PodListItemProps) {
  const t = useTranslations("mobile.access");
  const status = statusColors[pod.status] || statusColors.terminated;

  return (
    <SidebarPodContextMenu
      pod={pod}
      onRename={onRename}
      onShare={onShare}
      onOpenMobile={onOpenMobile}
      onPublishExpert={onPublishExpert}
      onTerminate={onTerminate}
      onDelete={onDelete}
      onWake={onWake}
      onTogglePerpetual={onTogglePerpetual}
    >
      <div
        data-testid="pod-list-item"
        data-pod-key={pod.pod_key}
        className={cn(
          "group flex items-center gap-2 px-3 py-2 motion-interactive hover:bg-surface-muted cursor-pointer",
          isOpen && "bg-muted/30"
        )}
        onClick={onClick}
      >
        <div className={cn("flex items-center justify-center", status.text)}>
          {getStatusIcon(pod.status)}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="text-sm truncate font-mono">
              {getPodDisplayName(pod)}
            </span>
            <AgentStatusBadge
              agentStatus={pod.agent_status ?? ''}
              podStatus={pod.status}
              variant="dot"
            />
            {isOpen && (
              <Terminal className="w-3 h-3 text-primary flex-shrink-0" />
            )}
          </div>
          {pod.created_by?.name && (
            <p className="text-xs text-muted-foreground truncate">
              {pod.created_by.name}
            </p>
          )}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="h-8 w-8 shrink-0 text-muted-foreground hover:text-foreground"
          title={t("mobileAccess")}
          aria-label={t("mobileAccess")}
          onClick={(event) => {
            event.stopPropagation();
            onOpenMobile();
          }}
        >
          <Smartphone className="h-4 w-4" />
        </Button>
        <SidebarPodActionsMenu
          pod={pod}
          onOpenMobile={onOpenMobile}
          onPublishExpert={onPublishExpert}
          onDelete={onDelete}
          onWake={onWake}
          onRename={onRename}
          onShare={onShare}
          onTerminate={onTerminate}
          onTogglePerpetual={onTogglePerpetual}
        />
      </div>
    </SidebarPodContextMenu>
  );
}

export { statusColors };
