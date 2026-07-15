import type {
  AgentArtifactItem,
  AgentConnectionStatus,
  AgentPlanStep,
  AgentSessionSnapshot,
  AgentTimelineItem,
  AgentConfigurationControl,
} from "@do-worker/agent-ui";
import { DEFAULT_AGENT_COMMANDS } from "@do-worker/agent-ui";

import type { AcpSessionState } from "@/stores/acpSessionTypes";
import { mergeWebAcpArtifacts } from "./webAcpArtifactProjection";
import { projectWebAcpPermission } from "./webAcpPermissionProjection";
import { projectWebAcpWorkspaceArtifacts } from "./webAcpWorkspaceArtifacts";

interface WebAcpSnapshotInput {
  agentLabel: string;
  connection: AgentConnectionStatus;
  sessionId: string;
  session: AcpSessionState;
  title: string;
  workspaceArtifacts?: AgentArtifactItem[];
}

interface TimedItem {
  timestamp: number;
  item: AgentTimelineItem;
}

export function projectWebAcpSnapshot({
  agentLabel,
  connection,
  sessionId,
  session,
  title,
  workspaceArtifacts = [],
}: WebAcpSnapshotInput): AgentSessionSnapshot {
  const supportsConfiguration =
    session.configuration.supportedPermissionModes.length > 0;
  return {
    sessionId,
    title,
    agentLabel,
    status: sessionStatus(session.state),
    connection,
    interactionMode: "acp",
    capabilities: {
      sendMessage: connection === "connected",
      interrupt: connection === "connected",
      resolvePermission: connection === "connected",
      updateConfiguration: connection === "connected" && supportsConfiguration,
      terminal: false,
    },
    commands: DEFAULT_AGENT_COMMANDS,
    configuration: supportsConfiguration ? configurationControls(session) : [],
    metadata: session.configuration.model
      ? [{ id: "model", label: "Model", value: session.configuration.model }]
      : [],
    items: mergeWebAcpArtifacts(timelineItems(session), workspaceArtifacts),
    plan: session.plan.map((step, index) => ({
      id: `plan-${index}-${step.title}`,
      title: step.title,
      status: planStatus(step.status),
    })),
    permissions: session.pendingPermissions.map(projectWebAcpPermission),
    terminals: [],
    hasOlderItems: false,
    error: null,
  };
}

function timelineItems(session: AcpSessionState): AgentTimelineItem[] {
  const timed: TimedItem[] = [];
  session.messages.forEach((message, index) => {
    const role =
      message.role === "user" || message.role === "system"
        ? message.role
        : "assistant";
    timed.push({
      timestamp: message.timestamp,
      item: {
        id: `message-${message.timestamp}-${index}`,
        kind: "message",
        role,
        text: message.text,
        status: message.complete === false ? "streaming" : "completed",
      },
    });
  });
  session.thinkings.forEach((thinking, index) => {
    timed.push({
      timestamp: thinking.timestamp,
      item: {
        id: `reasoning-${thinking.timestamp}-${index}`,
        kind: "reasoning",
        title: thinking.complete === false ? "Thinking" : "Reasoning",
        detail: thinking.text,
        status: thinking.complete === false ? "running" : "completed",
      },
    });
  });
  Object.values(session.toolCalls).forEach((tool) => {
    timed.push({
      timestamp: tool.timestamp,
      item: {
        id: `tool-${tool.toolCallId}`,
        kind: "tool",
        title: tool.toolName || "Tool",
        input: tool.argumentsJson,
        output: tool.resultText || tool.errorMessage,
        status: activityStatus(tool.status, tool.success),
      },
    });
    projectWebAcpWorkspaceArtifacts(tool).forEach((item) => {
      timed.push({ timestamp: tool.timestamp, item });
    });
  });
  session.logs.forEach((log, index) => {
    timed.push({
      timestamp: log.timestamp,
      item: {
        id: `log-${log.timestamp}-${index}`,
        kind: log.level === "error" ? "error" : "system",
        title: log.level === "error" ? "Agent error" : "Agent log",
        detail: log.message,
        status: log.level === "error" ? "failed" : "completed",
      },
    });
  });
  return timed.sort((left, right) => left.timestamp - right.timestamp).map(({ item }) => item);
}

function configurationControls(
  session: AcpSessionState,
): AgentConfigurationControl[] {
  const controls: AgentConfigurationControl[] = [];
  const modes = session.configuration.supportedPermissionModes;
  controls.push({
    id: "permissionMode",
    label: "Permissions",
    value: session.configuration.permissionMode,
    options: modes.map((value) => ({
      value,
      label: permissionModeLabel(value),
    })),
  });
  if (session.configuration.model) {
    controls.push({
      id: "model",
      label: "Model",
      value: session.configuration.model,
      options: [
        {
          value: session.configuration.model,
          label: session.configuration.model,
        },
      ],
    });
  }
  return controls;
}

function permissionModeLabel(value: string) {
  if (value === "bypass") return "Full access";
  if (value === "ask_dangerous") return "Ask for dangerous actions";
  if (value === "ask_any_write") return "Ask before writes";
  if (value === "bypassPermissions") return "Full access";
  if (value === "acceptEdits") return "Accept edits";
  if (value === "dontAsk") return "Do not ask";
  if (value === "default") return "Ask before changes";
  return value;
}

function sessionStatus(state: string): AgentSessionSnapshot["status"] {
  if (state === "processing") return "running";
  if (state === "waiting_permission") return "waiting";
  if (state === "error" || state === "failed") return "failed";
  return "idle";
}

function planStatus(status: string): AgentPlanStep["status"] {
  if (status === "in_progress") return "running";
  if (status === "completed") return "completed";
  if (status === "failed") return "failed";
  return "pending";
}

function activityStatus(
  status: string,
  success: boolean | undefined,
): "pending" | "running" | "completed" | "failed" {
  if (success === false || status === "failed" || status === "error") return "failed";
  if (success === true || status === "completed") return "completed";
  if (status === "in_progress" || status === "running") return "running";
  return "pending";
}
