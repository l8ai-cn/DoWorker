"use client";

import Link from "next/link";
import {
  Blocks,
  BookOpen,
  Bot,
  Code2,
  FolderGit2,
  Layers,
  MessageSquare,
  Network,
  Repeat,
  Server,
  Settings,
  Sparkles,
  Store,
  Terminal,
  Target,
  Ticket,
  Workflow,
  type LucideIcon,
} from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipPortal,
  TooltipTrigger,
} from "@radix-ui/react-tooltip";
import { cn } from "@/lib/utils";
import type { ActivityType } from "@/stores/ide";

const ICON_MAP: Record<string, LucideIcon> = {
  terminal: Terminal,
  ticket: Ticket,
  network: Network,
  "message-square": MessageSquare,
  target: Target,
  repeat: Repeat,
  bot: Bot,
  workflow: Workflow,
  blocks: Blocks,
  "book-open": BookOpen,
  code: Code2,
  repository: FolderGit2,
  server: Server,
  settings: Settings,
  layers: Layers,
  sparkles: Sparkles,
  store: Store,
};

interface ActivityBarLinkProps {
  id: ActivityType;
  icon: string;
  href: string;
  label: string;
  isActive: boolean;
  showBadge?: boolean;
  badgeCount?: number;
  onClick: (id: ActivityType) => void;
}

export function ActivityBarLink({
  id,
  icon,
  href,
  label,
  isActive,
  showBadge,
  badgeCount = 0,
  onClick,
}: ActivityBarLinkProps) {
  const Icon = ICON_MAP[icon] || Terminal;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={href}
          className={cn(
            "motion-interactive pressable relative flex h-9 w-full items-center gap-2 rounded-lg px-2.5",
            isActive
              ? "bg-surface-raised text-foreground shadow-[var(--shadow-soft)] ring-1 ring-border/45 before:absolute before:bottom-2 before:left-0 before:top-2 before:w-0.5 before:rounded-full before:bg-primary"
              : "text-muted-foreground hover:bg-surface-muted hover:text-foreground",
          )}
          onClick={() => onClick(id)}
        >
          <div className="relative shrink-0">
            <Icon className="h-4 w-4" />
            {showBadge && (
              <span className="absolute -right-2 -top-1.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-0.5 text-[9px] font-bold leading-none text-destructive-foreground">
                {badgeCount > 99 ? "99+" : badgeCount}
              </span>
            )}
          </div>
          <span className="truncate text-xs font-medium leading-tight">{label}</span>
        </Link>
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent
          side="right"
          className="z-50 rounded-md bg-popover px-2 py-1 text-sm text-popover-foreground shadow-[var(--shadow-soft)]"
        >
          {label}
        </TooltipContent>
      </TooltipPortal>
    </Tooltip>
  );
}
