import { clone, create } from "@bufbuild/protobuf";

import { AgentErrorSchema } from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import { UnsupportedValueSchema } from "@agent-cloud/proto/agent_workbench/v2/content_pb";
import { SessionConfigurationSchema } from "@agent-cloud/proto/agent_workbench/v2/configuration_pb";
import {
  EventEnvelopeSchema,
  PermissionRequestSchema,
  PermissionResolutionSchema,
  type AgentEvent,
  type SessionSnapshot,
  SessionResourceSchema,
  SupportCapabilitiesSchema,
  TerminalLeaseSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { PermissionRequestState } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import { AgentSessionReductionError } from "./agentSessionState";
import { upsertCommandReceipt } from "./commandReceiptTransitions";
import { mergeArtifactDescriptor } from "./mergeArtifactDescriptor";

export function applyAgentEvent(snapshot: SessionSnapshot, event: AgentEvent): void {
  const envelope = event.envelope;
  if (!envelope) throw new AgentSessionReductionError("event_envelope_missing");

  switch (event.event.case) {
    case "timelineItemAppended": {
      const content = event.event.value.content;
      if (!content) throw new AgentSessionReductionError("timeline_content_missing");
      if (snapshot.history.some((item) => item.envelope?.itemId === envelope.itemId)) {
        throw new AgentSessionReductionError("timeline_item_conflict");
      }
      snapshot.history.push(
        create(TimelineItemSchema, {
          envelope: clone(EventEnvelopeSchema, envelope),
          content: clone(TimelineItemContentSchema, content),
        }),
      );
      return;
    }
    case "timelineItemUpdated": {
      const content = event.event.value.content;
      const index = snapshot.history.findIndex(
        (item) => item.envelope?.itemId === envelope.itemId,
      );
      if (index < 0) throw new AgentSessionReductionError("timeline_item_missing");
      if (!content) throw new AgentSessionReductionError("timeline_content_missing");
      snapshot.history[index] = create(TimelineItemSchema, {
        envelope: clone(EventEnvelopeSchema, envelope),
        content: clone(TimelineItemContentSchema, content),
      });
      return;
    }
    case "commandReceiptChanged": {
      const receipt = event.event.value.receipt;
      if (!receipt) throw new AgentSessionReductionError("receipt_missing");
      snapshot.commandReceipts = upsertCommandReceipt(
        snapshot.commandReceipts,
        receipt,
        snapshot.sessionId,
      );
      return;
    }
    case "permissionRequested": {
      const request = event.event.value.request;
      if (!request) throw new AgentSessionReductionError("permission_request_missing");
      snapshot.permissionRequests = upsert(
        snapshot.permissionRequests,
        clone(PermissionRequestSchema, request),
        "permissionRequestId",
      );
      return;
    }
    case "permissionResolved": {
      const resolution = event.event.value.resolution;
      if (!resolution) throw new AgentSessionReductionError("permission_resolution_missing");
      const request = snapshot.permissionRequests.find(
        (item) => item.permissionRequestId === resolution.permissionRequestId,
      );
      if (!request) throw new AgentSessionReductionError("permission_request_missing");
      request.state = PermissionRequestState.RESOLVED;
      request.resolution = clone(PermissionResolutionSchema, resolution);
      return;
    }
    case "resourceChanged": {
      const resource = event.event.value.resource;
      if (!resource) throw new AgentSessionReductionError("resource_missing");
      snapshot.resources = upsert(
        snapshot.resources,
        clone(SessionResourceSchema, resource),
        "resourceId",
      );
      return;
    }
    case "artifactChanged": {
      const artifact = event.event.value.artifact;
      if (!artifact) throw new AgentSessionReductionError("artifact_missing");
      const index = snapshot.artifacts.findIndex(
        (item) => item.artifactId === artifact.artifactId,
      );
      const merged = mergeArtifactDescriptor(snapshot.artifacts[index], artifact);
      if (index < 0) snapshot.artifacts.push(merged);
      else snapshot.artifacts[index] = merged;
      return;
    }
    case "terminalLeaseChanged": {
      const { lease, resourceId } = event.event.value;
      const resource = snapshot.resources.find(
        (item) => item.resourceId === resourceId,
      );
      if (!resource || resource.resource.case !== "terminal") {
        throw new AgentSessionReductionError("terminal_resource_missing");
      }
      resource.resource.value.lease = lease
        ? clone(TerminalLeaseSchema, lease)
        : undefined;
      return;
    }
    case "capabilitiesChanged":
      snapshot.capabilities = event.event.value.capabilities
        ? clone(SupportCapabilitiesSchema, event.event.value.capabilities)
        : undefined;
      return;
    case "sessionStatusChanged":
      snapshot.status = event.event.value.status;
      snapshot.error = event.event.value.error
        ? clone(AgentErrorSchema, event.event.value.error)
        : undefined;
      return;
    case "configurationChanged":
      snapshot.configuration = event.event.value.configuration
        ? clone(SessionConfigurationSchema, event.event.value.configuration)
        : undefined;
      return;
    case "unsupported":
      if (snapshot.history.some((item) => item.envelope?.itemId === envelope.itemId)) {
        throw new AgentSessionReductionError("timeline_item_conflict");
      }
      snapshot.history.push(
        create(TimelineItemSchema, {
          envelope: clone(EventEnvelopeSchema, envelope),
          content: create(TimelineItemContentSchema, {
            content: {
              case: "unsupported",
              value: clone(UnsupportedValueSchema, event.event.value),
            },
          }),
        }),
      );
      return;
    default:
      throw new AgentSessionReductionError("event_payload_missing");
  }
}

function upsert<T extends object, K extends keyof T>(
  items: readonly T[],
  value: T,
  key: K,
): T[] {
  const next = items.slice();
  const index = next.findIndex((item) => item[key] === value[key]);
  if (index < 0) next.push(value);
  else next[index] = value;
  return next;
}
