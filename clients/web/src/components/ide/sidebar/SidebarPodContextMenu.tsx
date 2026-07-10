"use client";

import { useTranslations } from "next-intl";
import { Pencil, Share2, Square, RefreshCw, Bot, Smartphone } from "lucide-react";
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
  onTerminate: () => void;
  onTogglePerpetual: (perpetual: boolean) => void;
  children: React.ReactNode;
}

export function SidebarPodContextMenu({
  pod,
  onRename,
  onShare,
  onOpenMobile,
  onPublishExpert,
  onTerminate,
  onTogglePerpetual,
  children,
}: SidebarPodContextMenuProps) {
  const t = useTranslations("workspace");
  const tExpert = useTranslations("experts.publish");
  const isActive = pod.status === "running" || pod.status === "initializing";

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

        <ContextMenuSeparator />

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
