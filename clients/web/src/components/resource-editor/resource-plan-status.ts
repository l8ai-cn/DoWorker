import {
  PlanStatus,
  type ResourcePlan,
} from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";

export type ResourcePlanStatusKey =
  | "pending"
  | "applied"
  | "cancelled"
  | "expired"
  | "unavailable";

export function isPendingResourcePlan(
  plan: ResourcePlan | undefined,
): boolean {
  return plan?.status === PlanStatus.PENDING;
}

export function resourcePlanStatusKey(
  status: PlanStatus,
): ResourcePlanStatusKey {
  switch (status) {
    case PlanStatus.PENDING:
      return "pending";
    case PlanStatus.APPLIED:
      return "applied";
    case PlanStatus.CANCELLED:
      return "cancelled";
    case PlanStatus.EXPIRED:
      return "expired";
    case PlanStatus.UNSPECIFIED:
    default:
      return "unavailable";
  }
}
