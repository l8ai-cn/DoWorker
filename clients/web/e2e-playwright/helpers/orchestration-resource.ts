import {
  IssueSeverity,
  SourceFormat,
} from "../../../../proto/gen/ts/orchestration_resource/v1/orchestration_resource_types_pb";
import type { ConnectClient } from "./connect-client";

export type ApplicableResourceKind =
  | "ModelBinding"
  | "Repository"
  | "Skill"
  | "KnowledgeBase"
  | "EnvironmentBundle"
  | "ComputeTarget"
  | "ResourceProfile"
  | "ToolBinding"
  | "WorkerTemplate"
  | "Worker"
  | "Prompt"
  | "Expert"
  | "Workflow"
  | "GoalLoop";

interface PlanIssue {
  severity: IssueSeverity;
  path: string;
  code: string;
  message: string;
}

interface PlanResponse {
  issues: PlanIssue[];
  plan?: { planId: string };
}

export async function validatePlanApplyResource(
  client: ConnectClient,
  orgSlug: string,
  kind: ApplicableResourceKind,
  yaml: string,
): Promise<unknown> {
  const source = {
    format: SourceFormat.YAML,
    content: new TextEncoder().encode(yaml),
  };
  const validation = await client.orchestrationResource.validateResource({
    orgSlug,
    source,
  }) as PlanResponse;
  rejectBlockingIssues("Validate", validation.issues);

  const planned = await client.orchestrationResource.planResource({
    orgSlug,
    source,
  }) as PlanResponse;
  rejectBlockingIssues("Plan", planned.issues);
  const planId = planned.plan?.planId;
  if (!planId) {
    throw new Error(`Plan did not return a plan ID for ${kind}`);
  }
  return applyResourcePlan(client, orgSlug, kind, planId);
}

function applyResourcePlan(
  client: ConnectClient,
  orgSlug: string,
  kind: ApplicableResourceKind,
  planId: string,
) {
  const input = { orgSlug, planId };
  switch (kind) {
    case "ModelBinding":
    case "Repository":
    case "Skill":
    case "KnowledgeBase":
    case "EnvironmentBundle":
    case "ComputeTarget":
    case "ResourceProfile":
    case "ToolBinding":
      return client.orchestrationResource.applyBindingResourcePlan(input);
    case "WorkerTemplate":
      return client.orchestrationResource.applyWorkerTemplatePlan(input);
    case "Worker":
      return client.orchestrationResource.createWorkerFromPlan(input);
    case "Prompt":
      return client.orchestrationResource.applyPromptPlan(input);
    case "Expert":
      return client.orchestrationResource.applyExpertPlan(input);
    case "Workflow":
      return client.orchestrationResource.applyWorkflowPlan(input);
    case "GoalLoop":
      return client.orchestrationResource.createGoalLoopFromPlan(input);
    default:
      throw new Error(`Unsupported resource kind: ${kind satisfies never}`);
  }
}

function rejectBlockingIssues(stage: string, issues: PlanIssue[] = []) {
  const blocking = issues.filter(
    (issue) => issue.severity === IssueSeverity.BLOCKING,
  );
  if (blocking.length === 0) return;
  throw new Error(`${stage} blocked: ${blocking.map(
    (issue) => `${issue.path || "/"} ${issue.code}: ${issue.message}`,
  ).join("; ")}`);
}
