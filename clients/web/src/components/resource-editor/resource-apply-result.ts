import type {
  ApplyExpertPlanResponse,
  ApplyWorkerTemplatePlanResponse,
  ApplyWorkflowPlanResponse,
  CreateGoalLoopFromPlanResponse,
  CreateWorkerFromPlanResponse,
  Resource,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";

export type ResourceApplyResult =
  | Resource
  | ApplyExpertPlanResponse
  | ApplyWorkerTemplatePlanResponse
  | ApplyWorkflowPlanResponse
  | CreateGoalLoopFromPlanResponse
  | CreateWorkerFromPlanResponse;

export function appliedResource(
  result: ResourceApplyResult,
): Resource | undefined {
  if (result.$typeName === "proto.orchestration_resource.v1.Resource") {
    return result;
  }
  return result.resource;
}
