import {
  applyBindingResourcePlan,
  applyExpertPlan,
  applyPromptPlan,
  applyWorkerTemplatePlan,
  applyWorkflowPlan,
  createGoalLoopFromPlan,
  createWorkerFromPlan,
} from "@/lib/api/facade/orchestrationResource";
import {
  isResourceBindingKind,
  type ResourceEditorKind,
} from "./resource-editor-types";

export function applyResourcePlan(
  orgSlug: string,
  kind: ResourceEditorKind,
  planId: string,
) {
  if (isResourceBindingKind(kind)) {
    return applyBindingResourcePlan(orgSlug, planId);
  }
  switch (kind) {
    case "WorkerTemplate":
      return applyWorkerTemplatePlan(orgSlug, planId);
    case "Worker":
      return createWorkerFromPlan(orgSlug, planId);
    case "Prompt":
      return applyPromptPlan(orgSlug, planId);
    case "Expert":
      return applyExpertPlan(orgSlug, planId);
    case "Workflow":
      return applyWorkflowPlan(orgSlug, planId);
    case "GoalLoop":
      return createGoalLoopFromPlan(orgSlug, planId);
    default:
      throw new Error(`Unsupported resource kind: ${kind}`);
  }
}
