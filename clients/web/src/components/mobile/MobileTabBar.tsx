"use client";

import React from "react";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  useIDEStore,
  getMobileActivities,
  type ActivityType,
} from "@/stores/ide";
import { useCurrentOrg } from "@/stores/auth";
import { useTranslations } from "next-intl";
import {
  Terminal,
  Ticket,
  Network,
  MessageSquare,
  FolderGit2,
  Target,
  MoreHorizontal,
  Smartphone,
  type LucideIcon,
} from "lucide-react";

const ICON_MAP: Record<string, LucideIcon> = {
  terminal: Terminal,
  ticket: Ticket,
  network: Network,
  "message-square": MessageSquare,
  target: Target,
  repository: FolderGit2,
};

interface MobileTabBarProps {
  className?: string;
}

export function MobileTabBar({ className }: MobileTabBarProps) {
  const { activeActivity, setActiveActivity, setMobileMoreMenuOpen } =
    useIDEStore();
  const currentOrg = useCurrentOrg();
  const params = useParams();
  const pathname = usePathname();
  const t = useTranslations();
  const orgSlug = currentOrg?.slug || (params.org as string) || "";

  const mobileActivities = getMobileActivities();

  const getActivityRoute = (activity: ActivityType): string => {
    switch (activity) {
      case "workspace":
        return `/${orgSlug}/mobile/workers`;
      case "tickets":
        return `/${orgSlug}/tickets`;
      case "channels":
        return `/${orgSlug}/channels`;
      case "mesh":
        return `/${orgSlug}/mesh`;
      case "loops":
        return `/${orgSlug}/loops`;
      case "workflows":
        return `/${orgSlug}/workflows`;
      case "repositories":
        return `/${orgSlug}/repositories`;
      default:
        return `/${orgSlug}/workspace`;
    }
  };

  return (
    <nav
      className={cn(
        "h-16 bg-background panel-lift flex items-stretch px-2 safe-area-pb",
        className
      )}
    >
      {mobileActivities.map((activity) => {
        const Icon = activity.id === "workspace"
          ? Smartphone
          : ICON_MAP[activity.icon] || Terminal;
        const isActive = activity.id === "workspace"
          ? pathname.includes("/mobile/workers")
          : activeActivity === activity.id;

        return (
          <Link
            key={activity.id}
            href={getActivityRoute(activity.id)}
            className={cn(
              "flex-1 flex flex-col items-center justify-center gap-1 transition-colors",
              isActive
                ? "text-primary"
                : "text-muted-foreground active:text-foreground"
            )}
            onClick={() => setActiveActivity(activity.id)}
          >
            <Icon className="w-5 h-5" />
            <span className="text-[10px] font-medium">
              {activity.id === "workspace"
                ? t("mobile.workers.tab")
                : t(`ide.activities.${activity.id}`)}
            </span>
          </Link>
        );
      })}

      {/* More button */}
      <button
        className={cn(
          "flex-1 flex flex-col items-center justify-center gap-1 transition-colors",
          "text-muted-foreground active:text-foreground"
        )}
        onClick={() => setMobileMoreMenuOpen(true)}
      >
        <MoreHorizontal className="w-5 h-5" />
        <span className="text-[10px] font-medium">{t("mobile.more")}</span>
      </button>
    </nav>
  );
}

export default MobileTabBar;
