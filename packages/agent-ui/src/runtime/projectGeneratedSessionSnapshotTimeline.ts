import { MessageRole } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import type { PermissionRequest, TimelineItem } from "@do-worker/proto/agent_workbench/v2/session_pb";
import type { AgentPlanStep, AgentTimelineItem } from "../contracts";
import { projectArtifactReference, type ArtifactCatalog } from "./projectGeneratedSessionSnapshotArtifacts";
import { projectTimelineContent } from "./projectGeneratedSessionSnapshotContent";
import {
  formatAgentErrorDetail,
  formatUnsupported,
} from "./projectGeneratedSessionSnapshotPayload";
import {
  projectActivityStatus,
  projectMessageStatus,
  projectPermissionActivityStatus,
  sessionStatusLabel,
} from "./projectGeneratedSessionSnapshotStatuses";
import {
  messageRole,
  permissionTitle,
  planDetail,
  projectPlanStep,
  statusActivity,
  unsupportedTimeline,
} from "./projectGeneratedSessionSnapshotTimelineHelpers";
import { projectToolExecution } from "./projectGeneratedSessionSnapshotTool";
export interface TimelineProjection {
  items: AgentTimelineItem[];
  latestUserCommandId?: string;
  plan: AgentPlanStep[];
}

export function projectTimeline(
  history: readonly TimelineItem[],
  catalog: ArtifactCatalog,
  permissions: ReadonlyMap<string, PermissionRequest>,
): TimelineProjection {
  const projection: TimelineProjection = { items: [], plan: [] };
  history.forEach((timelineItem, index) => {
    const id = timelineItem.envelope?.itemId || `timeline:${index}`;
    const content = timelineItem.content?.content;
    if (!content?.case) {
      projection.items.push(unsupportedTimeline(id, "missing timeline content"));
      return;
    }
    if (content.case === "message") {
      if (
        content.value.role === MessageRole.USER &&
        timelineItem.envelope?.causationCommandId
      ) {
        projection.latestUserCommandId =
          timelineItem.envelope.causationCommandId;
      }
      const blocks = projectTimelineContent(content.value.content, id, catalog);
      if (blocks.text.length > 0) {
        projection.items.push({
          id,
          kind: "message",
          role: messageRole(content.value.role),
          text: blocks.text.join("\n\n"),
          status: projectMessageStatus(content.value.status),
        });
      }
      projection.items.push(...blocks.attachments, ...blocks.artifacts, ...blocks.evidence);
      return;
    }
    if (content.case === "reasoning") {
      const blocks = projectTimelineContent(content.value.content, id, catalog);
      projection.items.push({
        id,
        kind: "reasoning",
        title: "Reasoning",
        detail: blocks.text.join("\n\n") || undefined,
        status: projectActivityStatus(content.value.status),
      });
      projection.items.push(...blocks.attachments, ...blocks.artifacts, ...blocks.evidence);
      return;
    }
    if (content.case === "toolExecution") {
      projection.items.push(...projectToolExecution(content.value, id, catalog));
      return;
    }
    if (content.case === "plan") {
      projection.plan = content.value.steps.map(projectPlanStep);
      projection.items.push({
        id,
        kind: "system",
        title: "Plan updated",
        detail: planDetail(content.value.steps),
        status: "completed",
      });
      return;
    }
    if (content.case === "artifactReference") {
      projection.items.push(
        projectArtifactReference(content.value.artifact, id, catalog),
      );
      return;
    }
    if (content.case === "approval") {
      const request = permissions.get(content.value.permissionRequestId);
      projection.items.push({
        id,
        kind: "system",
        title: "Permission requested",
        detail: [
          `permissionRequestId=${content.value.permissionRequestId}`,
          permissionTitle(request),
        ]
          .filter(Boolean)
          .join("\n"),
        status: projectPermissionActivityStatus(request),
      });
      return;
    }
    if (content.case === "status") {
      projection.items.push({
        id,
        kind: "system",
        title: `Session ${sessionStatusLabel(content.value.status)}`,
        detail: content.value.detail,
        status: statusActivity(content.value.status),
      });
      return;
    }
    if (content.case === "error") {
      const error = content.value.error;
      projection.items.push({
        id,
        kind: "error",
        title: error?.code || "agent_error",
        detail: error ? formatAgentErrorDetail(error) : "missing error payload",
        status: "failed",
      });
      return;
    }
    if (content.case === "system") {
      const blocks = projectTimelineContent(content.value.content, id, catalog);
      projection.items.push({
        id,
        kind: "system",
        title: "System",
        detail: blocks.text.join("\n\n") || undefined,
        status: "completed",
      });
      projection.items.push(...blocks.attachments, ...blocks.artifacts, ...blocks.evidence);
      return;
    }
    projection.items.push(unsupportedTimeline(id, formatUnsupported(content.value)));
  });
  return projection;
}
