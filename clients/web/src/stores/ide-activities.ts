import type { ActivityType } from "./ide";

export interface ActivityConfig {
  id: ActivityType;
  label: string;
  icon: string;
  group: "comm" | "build" | "ops" | "system";
  mobileVisible: boolean;
  mobileOrder?: number;
}

export const ACTIVITIES: ActivityConfig[] = [
  { id: "channels", label: "Channels", icon: "message-square", group: "comm", mobileVisible: true, mobileOrder: 1 },
  { id: "mesh", label: "Mesh", icon: "network", group: "comm", mobileVisible: true, mobileOrder: 2 },
  { id: "workspace", label: "Workspace", icon: "terminal", group: "build", mobileVisible: true, mobileOrder: 3 },
  { id: "tickets", label: "Tickets", icon: "ticket", group: "build", mobileVisible: true, mobileOrder: 4 },
  { id: "loops", label: "Loops", icon: "target", group: "build", mobileVisible: false },
  { id: "workflows", label: "Workflows", icon: "repeat", group: "build", mobileVisible: false },
  { id: "experts", label: "Experts", icon: "bot", group: "build", mobileVisible: false },
  { id: "automation", label: "Automation", icon: "workflow", group: "build", mobileVisible: false },
  { id: "apiAccess", label: "API Access", icon: "code", group: "build", mobileVisible: false },
  { id: "knowledge", label: "Knowledge Base", icon: "book-open", group: "ops", mobileVisible: false },
  { id: "infra", label: "Infra", icon: "layers", group: "ops", mobileVisible: false },
  { id: "marketplace", label: "Marketplace", icon: "store", group: "ops", mobileVisible: false },
  { id: "skills", label: "Skills", icon: "sparkles", group: "ops", mobileVisible: false },
  { id: "settings", label: "Settings", icon: "settings", group: "system", mobileVisible: false },
];

export function getMobileActivities(): ActivityConfig[] {
  return ACTIVITIES.filter((activity) => activity.mobileVisible).sort(
    (left, right) => (left.mobileOrder ?? 99) - (right.mobileOrder ?? 99),
  );
}

export function getMoreMenuActivities(): ActivityConfig[] {
  return ACTIVITIES.filter((activity) => !activity.mobileVisible);
}
