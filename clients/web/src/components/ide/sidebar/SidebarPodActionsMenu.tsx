"use client";

import { Bot, MoreHorizontal, Pencil, RefreshCw, Share2, Smartphone, Square, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Pod } from "@/stores/pod";

interface SidebarPodActionsMenuProps {
  pod: Pod;
  onRename: () => void;
  onShare: () => void;
  onOpenMobile: () => void;
  onPublishExpert?: () => void;
  onDelete: () => void;
  onTerminate: () => void;
  onTogglePerpetual: (perpetual: boolean) => void;
}

export function SidebarPodActionsMenu({
  pod,
  onRename,
  onShare,
  onOpenMobile,
  onPublishExpert,
  onDelete,
  onTerminate,
  onTogglePerpetual,
}: SidebarPodActionsMenuProps) {
  const t = useTranslations("workspace");
  const tExpert = useTranslations("experts.publish");
  const isActive = pod.status === "running" || pod.status === "initializing";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          aria-label="Worker actions"
          className="h-7 w-7 shrink-0 p-0"
          onClick={(event) => event.stopPropagation()}
          size="sm"
          variant="ghost"
        >
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" onClick={(event) => event.stopPropagation()}>
        <DropdownMenuItem onClick={onRename}>
          <Pencil className="mr-2 h-4 w-4" />{t("contextMenu.rename")}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onShare}>
          <Share2 className="mr-2 h-4 w-4" />{t("contextMenu.share")}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onOpenMobile}>
          <Smartphone className="mr-2 h-4 w-4" />{t("contextMenu.mobileAccess")}
        </DropdownMenuItem>
        {onPublishExpert && (
          <DropdownMenuItem onClick={onPublishExpert}>
            <Bot className="mr-2 h-4 w-4" />{tExpert("contextMenu")}
          </DropdownMenuItem>
        )}
        {isActive && (
          <DropdownMenuItem onClick={() => onTogglePerpetual(!pod.perpetual)}>
            <RefreshCw className="mr-2 h-4 w-4" />
            {pod.perpetual ? t("contextMenu.disablePerpetual") : t("contextMenu.enablePerpetual")}
          </DropdownMenuItem>
        )}
        <DropdownMenuSeparator />
        {!isActive && (
          <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={onDelete}>
            <Trash2 className="mr-2 h-4 w-4" />{t("contextMenu.delete")}
          </DropdownMenuItem>
        )}
        <DropdownMenuItem className="text-destructive focus:text-destructive" disabled={!isActive} onClick={onTerminate}>
          <Square className="mr-2 h-4 w-4" />{t("contextMenu.terminate")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
