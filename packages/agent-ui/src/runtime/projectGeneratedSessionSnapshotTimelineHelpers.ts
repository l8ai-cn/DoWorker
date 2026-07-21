import { MessageRole } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import type { PermissionRequest, PlanStep } from "@agent-cloud/proto/agent_workbench/v2/session_pb";

import type { AgentActivityItem, AgentPlanStep } from "../contracts";
import {
  projectPlanStepStatus,
  projectSessionStatus,
} from "./projectGeneratedSessionSnapshotStatuses";

export function projectPlanStep(step: PlanStep): AgentPlanStep {
  return {
    id: step.stepId,
    title: step.title,
    status: projectPlanStepStatus(step.status),
  };
}

export function planDetail(steps: readonly PlanStep[]): string {
  return steps
    .map(
      (step) =>
        `${step.title} [${projectPlanStepStatus(step.status)}]${step.detail ? `\n${step.detail}` : ""}`,
    )
    .join("\n");
}

export function messageRole(role: MessageRole): "user" | "assistant" | "system" {
  if (role === MessageRole.USER) return "user";
  if (role === MessageRole.ASSISTANT) return "assistant";
  return "system";
}

export function statusActivity(
  status: Parameters<typeof projectSessionStatus>[0],
): AgentActivityItem["status"] {
  const projected = projectSessionStatus(status);
  if (
    projected === "launching" ||
    projected === "running" ||
    projected === "waiting"
  ) {
    return "running";
  }
  if (projected === "failed") return "failed";
  return "completed";
}

export function permissionTitle(request: PermissionRequest | undefined): string {
  if (request?.request.case === "approval") return request.request.value.title;
  if (request?.request.case === "questionnaire") return request.request.value.title;
  return "";
}

export function unsupportedTimeline(id: string, detail: string): AgentActivityItem {
  return {
    id,
    kind: "system",
    title: "Unsupported timeline item",
    detail,
    status: "failed",
  };
}
