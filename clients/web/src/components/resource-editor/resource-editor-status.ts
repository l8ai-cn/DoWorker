import type { BadgeProps } from "@/components/ui/badge";
import type { ResourceDraftState } from "./resource-draft-state-types";
import { resourcePlanStatusKey } from "./resource-plan-status";

type EditorStatusKey =
  | "status.draft"
  | "status.planReady"
  | "status.planApplied"
  | "status.planCancelled"
  | "status.planExpired"
  | "status.planUnavailable"
  | "status.applied";

interface EditorStatus {
  key: EditorStatusKey;
  variant: BadgeProps["variant"];
}

export function resourceEditorStatus(
  state: ResourceDraftState,
): EditorStatus {
  if (state.apply.status === "ready") {
    return { key: "status.applied", variant: "success" };
  }
  if (state.plan.status === "expired") {
    return { key: "status.planExpired", variant: "warning" };
  }
  if (state.plan.status !== "ready" || !state.plan.response.plan) {
    return { key: "status.draft", variant: "secondary" };
  }
  switch (resourcePlanStatusKey(state.plan.response.plan.status)) {
    case "pending":
      return { key: "status.planReady", variant: "info" };
    case "applied":
      return { key: "status.planApplied", variant: "success" };
    case "cancelled":
      return { key: "status.planCancelled", variant: "warning" };
    case "expired":
      return { key: "status.planExpired", variant: "warning" };
    case "unavailable":
      return { key: "status.planUnavailable", variant: "destructive" };
  }
}
