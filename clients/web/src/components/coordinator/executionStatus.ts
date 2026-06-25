import type { BadgeProps } from "@/components/ui/badge";

export const EXECUTION_COLUMNS = ["pending", "running", "succeeded", "failed"] as const;

export type ExecutionColumn = (typeof EXECUTION_COLUMNS)[number];

// Collapse the backend's finer-grained statuses (claimed, cancelled,
// feedback_failed) onto the four board columns the UI renders.
export function statusColumn(status: string): ExecutionColumn {
  switch (status) {
    case "pending":
    case "claimed":
      return "pending";
    case "running":
      return "running";
    case "succeeded":
      return "succeeded";
    default:
      return "failed";
  }
}

export function statusBadgeVariant(status: string): BadgeProps["variant"] {
  switch (statusColumn(status)) {
    case "succeeded":
      return "default";
    case "failed":
      return "destructive";
    case "running":
      return "secondary";
    default:
      return "outline";
  }
}
