// Goal status → status-bar tone. Mirrors loopal TUI's goal indicator colors
// (active / paused / complete / infeasible).
const GOAL_STATUS_TONE: Record<string, string> = {
  active: "text-success",
  paused: "text-warning",
  complete: "text-info",
  infeasible: "text-danger",
};

export function goalStatusTone(status: string): string {
  return GOAL_STATUS_TONE[status] ?? "text-muted-foreground";
}
