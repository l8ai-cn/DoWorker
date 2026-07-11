"use client";

import { useTranslations } from "next-intl";
import { Pencil, Share2, Square, RefreshCw, RotateCcw, Bot, Smartphone, Trash2 } from "lucide-react";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import type { Pod } from "@/stores/pod";

interface SidebarPodContextMenuProps {
  pod: Pod;
  onRename: () => void;
  onShare: () => void;
  onOpenMobile: () => void;
  onPublishExpert?: () => void;
  onDelete: () => void;
  onTerminate: () => void;
  onWake: () => void;
  onTogglePerpetual: (perpetual: boolean) => void;
  children: React.ReactNode;
}

export function SidebarPodContextMenu({
  pod,
  onRename,
  onShare,
  onOpenMobile,
  onPublishExpert,
  onDelete,
  onTerminate,
  onWake,
  onTogglePerpetual,
  children,
}: SidebarPodContextMenuProps) {
  const t = useTranslations("workspace");
  const tExpert = useTranslations("experts.publish");
  const isActive = pod.status === "running" || pod.status === "initializing";
  const isWakeable = pod.status === "terminated" || pod.status === "completed";

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={onRename}>
          <Pencil className="mr-2 h-4 w-4" />
          {t("contextMenu.rename")}
        </ContextMenuItem>
        <ContextMenuItem onClick={onShare}>
          <Share2 className="mr-2 h-4 w-4" />
          {t("contextMenu.share")}
        </ContextMenuItem>
        <ContextMenuItem onClick={onOpenMobile}>
          <Smartphone className="mr-2 h-4 w-4" />
          {t("contextMenu.mobileAccess")}
        </ContextMenuItem>
        {onPublishExpert && (
          <ContextMenuItem onClick={onPublishExpert}>
            <Bot className="mr-2 h-4 w-4" />
            {tExpert("contextMenu")}
          </ContextMenuItem>
        )}

        {isActive && (
          <ContextMenuItem onClick={() => onTogglePerpetual(!pod.perpetual)}>
            <RefreshCw className="mr-2 h-4 w-4" />
            {pod.perpetual
              ? t("contextMenu.disablePerpetual")
              : t("contextMenu.enablePerpetual")}
          </ContextMenuItem>
        )}
        {isWakeable && (
          <ContextMenuItem onClick={onWake}>
            <RotateCcw className="mr-2 h-4 w-4" />
            {t("contextMenu.wake")}
          </ContextMenuItem>
        )}

        <ContextMenuSeparator />
        {!isActive && (
          <ContextMenuItem onClick={onDelete} className="text-destructive focus:text-destructive">
            <Trash2 className="mr-2 h-4 w-4" />
            {t("contextMenu.delete")}
          </ContextMenuItem>
        )}

        <ContextMenuItem
          onClick={onTerminate}
          disabled={!isActive}
          className="text-destructive focus:text-destructive"
        >
          <Square className="mr-2 h-4 w-4" />
          {t("contextMenu.terminate")}
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}
