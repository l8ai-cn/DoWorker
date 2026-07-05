import type { ActivityType } from "@/stores/ide";

const ACTIVITIES_WITH_SIDEBAR = new Set<ActivityType>([
  "workspace",
  "tickets",
  "channels",
  "mesh",
  "loops",
  "blocks",
  "infra",
  "repositories",
  "runners",
  "skills",
  "settings",
]);

export function activityHasSidebar(activity: ActivityType): boolean {
  return ACTIVITIES_WITH_SIDEBAR.has(activity);
}
