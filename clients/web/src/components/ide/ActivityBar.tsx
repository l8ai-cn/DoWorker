"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useParams } from "next/navigation";
import {
  Tooltip,
  TooltipContent,
  TooltipPortal,
  TooltipProvider,
  TooltipTrigger,
} from "@radix-ui/react-tooltip";
import { cn } from "@/lib/utils";
import { useIDEStore, ACTIVITIES, type ActivityType } from "@/stores/ide";
import { resolveActivityFromPathname } from "@/lib/ide-route";
import { useCurrentOrg } from "@/stores/auth";
import { useTotalUnreadCount } from "@/stores/channelMessageStore";
import { useTranslations } from "next-intl";
import { CircleHelp, ShieldCheck } from "lucide-react";
import { OrgSwitcher } from "@/components/ide/OrgSwitcher";
import { ReminderArea } from "@/components/ide/ReminderArea";
import { ActivityBarLink } from "./ActivityBarLink";
import { useIsSystemAdmin } from "@/hooks/useIsSystemAdmin";

interface ActivityBarProps {
  className?: string;
}

export function ActivityBar({ className }: ActivityBarProps) {
  const activeActivity = useIDEStore((s) => s.activeActivity);
  const setActiveActivity = useIDEStore((s) => s.setActiveActivity);
  const currentOrg = useCurrentOrg();
  const params = useParams();
  const pathname = usePathname();
  const orgSlug = currentOrg?.slug || (params.org as string) || "";
  const t = useTranslations();
  const totalChannelUnread = useTotalUnreadCount();
  const isSystemAdmin = useIsSystemAdmin();

  const getActivityRoute = (activity: ActivityType): string => {
    switch (activity) {
      case "workspace":
        return `/${orgSlug}/workspace`;
      case "tickets":
        return `/${orgSlug}/tickets`;
      case "channels":
        return `/${orgSlug}/channels`;
      case "mesh":
        return `/${orgSlug}/mesh`;
      case "loops":
        return `/${orgSlug}/loops`;
      case "experts":
        return `/${orgSlug}/experts`;
      case "automation":
        return `/${orgSlug}/automation`;
      case "apiAccess":
        return `/${orgSlug}/api-access`;
      case "knowledge":
        return `/${orgSlug}/knowledge-base`;
      case "blocks":
        return `/${orgSlug}/blocks`;
      case "infra":
        return `/${orgSlug}/infra?tab=runners`;
      case "repositories":
        return `/${orgSlug}/repositories`;
      case "runners":
        return `/${orgSlug}/runners`;
      case "skills":
        return `/${orgSlug}/skills`;
      case "settings":
        return `/${orgSlug}/settings`;
      default:
        return `/${orgSlug}/workspace`;
    }
  };

  React.useEffect(() => {
    const activity = resolveActivityFromPathname(pathname);
    if (activity) setActiveActivity(activity);
  }, [pathname, setActiveActivity]);

  const mainActivities = ACTIVITIES.filter((a) => a.id !== "settings");
  const bottomActivities = ACTIVITIES.filter((a) => a.id === "settings");

  return (
    <TooltipProvider delayDuration={300}>
      <aside
        className={cn(
          "w-[120px] bg-surface flex flex-col",
          className
        )}
      >
        <div className="flex h-14 items-center justify-start px-2.5">
          <OrgSwitcher />
        </div>

        <nav className="flex-1 flex flex-col items-stretch py-2 gap-1 px-2">
          {mainActivities.map((activity, idx) => {
            const isActive = activeActivity === activity.id;
            const showBadge = activity.id === "channels" && totalChannelUnread > 0;
            const prev = mainActivities[idx - 1];
            const showDivider = prev && prev.group !== activity.group;

            return (
              <React.Fragment key={activity.id}>
                {showDivider && (
                  <div className="my-1 h-2" aria-hidden="true" />
                )}
                <ActivityBarLink
                  id={activity.id}
                  icon={activity.icon}
                  href={getActivityRoute(activity.id)}
                  label={t(`ide.activities.${activity.id}`)}
                  isActive={isActive}
                  showBadge={showBadge}
                  badgeCount={totalChannelUnread}
                  onClick={setActiveActivity}
                />
              </React.Fragment>
            );
          })}
        </nav>

        <ReminderArea />

        <nav className="flex flex-col items-stretch py-2 gap-1 px-2 pt-3">
          <Tooltip>
            <TooltipTrigger asChild>
              <a
                href="https://discord.gg/3RcX7VBbH9"
                target="_blank"
                rel="noopener noreferrer"
                className="motion-interactive pressable w-full h-9 px-2.5 flex items-center gap-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-muted"
              >
                <CircleHelp className="w-4 h-4 shrink-0" />
                <span className="text-xs leading-tight font-medium truncate">
                  {t("ide.activities.help")}
                </span>
              </a>
            </TooltipTrigger>
            <TooltipPortal>
              <TooltipContent
                side="right"
                className="z-50 bg-popover text-popover-foreground px-2 py-1 text-sm rounded-md shadow-[var(--shadow-soft)]"
              >
                {t("ide.activities.help")}
              </TooltipContent>
            </TooltipPortal>
          </Tooltip>

          {isSystemAdmin && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Link
                  href="/admin/audit-logs"
                  className="motion-interactive pressable w-full h-9 px-2.5 flex items-center gap-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-muted"
                >
                  <ShieldCheck className="w-4 h-4 shrink-0" />
                  <span className="text-xs leading-tight font-medium truncate">
                    Admin
                  </span>
                </Link>
              </TooltipTrigger>
              <TooltipPortal>
                <TooltipContent
                  side="right"
                  className="z-50 bg-popover text-popover-foreground px-2 py-1 text-sm rounded-md shadow-[var(--shadow-soft)]"
                >
                  Admin
                </TooltipContent>
              </TooltipPortal>
            </Tooltip>
          )}

          {bottomActivities.map((activity) => {
            const isActive = activeActivity === activity.id;

            return (
              <ActivityBarLink
                key={activity.id}
                id={activity.id}
                icon={activity.icon}
                href={getActivityRoute(activity.id)}
                label={t(`ide.activities.${activity.id}`)}
                isActive={isActive}
                onClick={setActiveActivity}
              />
            );
          })}
        </nav>
      </aside>
    </TooltipProvider>
  );
}

export default ActivityBar;
