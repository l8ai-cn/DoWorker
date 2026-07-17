import { PermissionDecision } from "@do-worker/proto/agent_workbench/v2/command_pb";
import type { PermissionRequest } from "@do-worker/proto/agent_workbench/v2/session_pb";
import { ToolPhase } from "@do-worker/proto/agent_workbench/v2/tool_pb";
import {
  PermissionRequestState,
  PlanStepStatus,
  SessionResourceStatus,
  SessionStatus,
  TimelineItemStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";

import type {
  AgentActivityItem,
  AgentConnectionStatus,
  AgentMessageItem,
  AgentPlanStep,
  AgentSessionStatus,
  AgentToolStatus,
} from "../contracts";

export function projectSessionStatus(
  status: SessionStatus,
): AgentSessionStatus {
  if (status === SessionStatus.LAUNCHING) return "launching";
  if (status === SessionStatus.RUNNING) return "running";
  if (status === SessionStatus.WAITING) return "waiting";
  if (status === SessionStatus.COMPLETED) return "completed";
  if (
    status === SessionStatus.FAILED ||
    status === SessionStatus.RESYNC_REQUIRED
  ) {
    return "failed";
  }
  return "idle";
}

export function sessionStatusLabel(status: SessionStatus): string {
  if (status === SessionStatus.IDLE) return "idle";
  if (status === SessionStatus.LAUNCHING) return "launching";
  if (status === SessionStatus.RUNNING) return "running";
  if (status === SessionStatus.WAITING) return "waiting";
  if (status === SessionStatus.COMPLETED) return "completed";
  if (status === SessionStatus.FAILED) return "failed";
  if (status === SessionStatus.RESYNC_REQUIRED) return "resync required";
  return "unspecified";
}

export function projectMessageStatus(
  status: TimelineItemStatus,
): AgentMessageItem["status"] {
  if (status === TimelineItemStatus.COMPLETED) return "completed";
  if (
    status === TimelineItemStatus.FAILED ||
    status === TimelineItemStatus.CANCELLED ||
    status === TimelineItemStatus.UNSPECIFIED
  ) {
    return "failed";
  }
  return "streaming";
}

export function projectActivityStatus(
  status: TimelineItemStatus,
): AgentActivityItem["status"] {
  if (status === TimelineItemStatus.PENDING) return "pending";
  if (
    status === TimelineItemStatus.STREAMING ||
    status === TimelineItemStatus.RUNNING
  ) {
    return "running";
  }
  if (status === TimelineItemStatus.COMPLETED) return "completed";
  return "failed";
}

export function projectPlanStepStatus(
  status: PlanStepStatus,
): AgentPlanStep["status"] {
  if (status === PlanStepStatus.RUNNING) return "running";
  if (
    status === PlanStepStatus.COMPLETED ||
    status === PlanStepStatus.SKIPPED
  ) {
    return "completed";
  }
  if (status === PlanStepStatus.FAILED) return "failed";
  return "pending";
}

export function projectToolStatus(status: ToolPhase): AgentToolStatus {
  if (status === ToolPhase.QUEUED) return "pending";
  if (
    status === ToolPhase.RUNNING ||
    status === ToolPhase.WAITING_APPROVAL
  ) {
    return "running";
  }
  if (status === ToolPhase.COMPLETED) return "completed";
  return "failed";
}

export function projectResourceStatus(
  status: SessionResourceStatus,
): AgentConnectionStatus {
  if (status === SessionResourceStatus.CONNECTING) return "connecting";
  if (status === SessionResourceStatus.READY) return "connected";
  if (status === SessionResourceStatus.DEGRADED) return "reconnecting";
  return "disconnected";
}

export function projectPermissionActivityStatus(
  request: PermissionRequest | undefined,
): AgentActivityItem["status"] {
  if (request?.state === PermissionRequestState.PENDING) return "pending";
  if (
    request?.state === PermissionRequestState.RESOLVED &&
    request.resolution?.decision === PermissionDecision.ACCEPT
  ) {
    return "completed";
  }
  return "failed";
}
