import { PermissionDecision } from "@do-worker/proto/agent_workbench/v2/command_pb";

import type {
  AgentArtifactActionCommand,
  AgentPermissionResolution,
} from "../contracts";
import type { CreateAgentCommandEnvelopeInput } from "./createAgentCommandEnvelope";

type CommandPayload = CreateAgentCommandEnvelopeInput["command"];

export function sendPromptPayload(text: string): CommandPayload {
  const normalized = text.trim();
  if (!normalized) throw new Error("agent_workbench_prompt_missing");
  return {
    case: "sendPrompt",
    value: { text: normalized, attachments: [] },
  };
}

export function interruptPayload(turnId?: string): CommandPayload {
  return {
    case: "interrupt",
    value: { turnId, reason: "user_requested" },
  };
}

export function configurationPayload(
  patch: Record<string, unknown>,
): CommandPayload {
  const values = Object.entries(patch).map(([key, value]) => ({
    key,
    value: structuredJson(value),
  }));
  if (values.length === 0) {
    throw new Error("agent_workbench_configuration_empty");
  }
  return { case: "changeConfiguration", value: { values } };
}

export function permissionPayload(
  permissionRequestId: string,
  resolution: AgentPermissionResolution,
): CommandPayload {
  return {
    case: "resolvePermission",
    value: {
      permissionRequestId,
      decision:
        resolution.action === "accept"
          ? PermissionDecision.ACCEPT
          : PermissionDecision.DECLINE,
      response:
        resolution.action === "accept"
          ? structuredJson(resolution.content)
          : undefined,
    },
  };
}

export function artifactActionPayload(
  command: AgentArtifactActionCommand,
): CommandPayload {
  if (!command.actionSchemaVersion.trim()) {
    throw new Error("agent_workbench_action_schema_missing");
  }
  return {
    case: "artifactAction",
    value: {
      artifactId: command.artifactId,
      representationId: command.representationId ?? "",
      baseRevision: command.baseRevision,
      clientActionId: command.commandId,
      actionType: command.actionType,
      payload: structuredJson(command.payload),
      actionSchemaVersion: command.actionSchemaVersion,
    },
  };
}

function structuredJson(value: unknown) {
  return {
    mediaType: "application/json",
    data: new TextEncoder().encode(
      JSON.stringify(value, (_, item: unknown) =>
        typeof item === "bigint" ? item.toString() : item,
      ),
    ),
  };
}
