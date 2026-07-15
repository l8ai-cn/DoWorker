import type {
  AgentPermissionRequest,
  AgentSessionMetadata,
  AgentSessionSnapshot,
  AgentTimelineItem,
} from "@do-worker/agent-ui";
import { DEFAULT_AGENT_COMMANDS, workspaceFileArtifacts } from "@do-worker/agent-ui";

import type { AnyBlock } from "@/lib/blocks";
import { buildBubbles } from "@/lib/renderItems";
import type { ActiveResponse } from "@/store/types";
import type { SessionStatus } from "@/lib/types";
import type { EmbeddedSession } from "@/embed-session-api";
import type { EmbeddedAcpConfiguration } from "./embeddedAcpRelayCodec";
import { projectEmbeddedPermission } from "./embeddedPermissionProjection";
import { projectEmbeddedTimelineItem } from "./embeddedTimelineProjection";

type ProjectionInput = Omit<
  AgentSessionSnapshot,
  | "agentLabel"
  | "configuration"
  | "interactionMode"
  | "items"
  | "permissions"
  | "plan"
  | "sessionId"
  | "title"
> & {
  activeResponse: ActiveResponse | null;
  blocks: AnyBlock[];
  configuration: EmbeddedAcpConfiguration;
  configurationConnected: boolean;
  session: EmbeddedSession | null;
  status: SessionStatus;
  workspaceArtifacts?: AgentTimelineItem[];
};

export function projectEmbeddedWorkspaceSnapshot(
  input: ProjectionInput,
): AgentSessionSnapshot {
  const items: AgentTimelineItem[] = [];
  const permissions: AgentPermissionRequest[] = [];
  const artifactIds = new Set<string>();
  const bubbles = buildBubbles(input.blocks, input.activeResponse);

  bubbles.forEach((bubble) => {
    if (bubble.kind === "user") {
      items.push({
        id: bubble.stableKey ?? bubble.itemId,
        kind: "message",
        role: "user",
        status: "completed",
        text: bubble.content
          .filter((content) => content.type === "input_text")
          .map((content) => content.text)
          .join(""),
      });
      return;
    }
    if (bubble.kind === "assistant") {
      bubble.items.forEach((item, itemIndex) => {
        const id = `${bubble.stableId}:${item.itemId ?? itemIndex}`;
        if (item.kind === "elicitation") {
          if (item.status === "pending") {
            permissions.push(projectEmbeddedPermission(item));
          }
          return;
        }
        items.push(projectEmbeddedTimelineItem(item, id, bubble.lifecycle));
        if (
          item.kind === "tool" &&
          item.execution.name === "fileChange" &&
          item.state === "output-available"
        ) {
          workspaceFileArtifacts(id, item.execution.arguments.changes).forEach((artifact) => {
            if (artifactIds.has(artifact.artifactId)) return;
            artifactIds.add(artifact.artifactId);
            items.push(artifact);
          });
        }
      });
      return;
    }
    if (bubble.kind === "routing_decision") {
      items.push({
        id: bubble.itemId,
        kind: "system",
        status: "completed",
        title: `Model selected: ${bubble.model}`,
        detail: bubble.rationale,
      });
      return;
    }
    items.push({
      id: bubble.itemId,
      kind: "system",
      status: bubble.kind === "compaction_loading" ? "running" : "completed",
      title:
        bubble.kind === "compaction_loading"
          ? "Compacting conversation"
          : "Conversation compacted",
    });
  });
  (input.workspaceArtifacts ?? []).forEach((artifact) => {
    if (artifact.kind !== "artifact" || artifactIds.has(artifact.artifactId)) return;
    artifactIds.add(artifact.artifactId);
    items.push(artifact);
  });

  const session = input.session;
  const hasTerminal = session?.interactionMode === "pty" && input.terminals.length > 0;
  const supportsConfiguration =
    input.configurationConnected &&
    input.configuration.supportedPermissionModes.length > 0;
  return {
    agentLabel: session?.agentLabel ?? "Agent",
    capabilities: {
      ...input.capabilities,
      terminal: hasTerminal,
      updateConfiguration: supportsConfiguration,
    },
    commands: input.capabilities.sendMessage ? DEFAULT_AGENT_COMMANDS : [],
    configuration: supportsConfiguration
      ? [
          {
            id: "permissionMode",
            label: "Permissions",
            value: input.configuration.permissionMode,
            options: input.configuration.supportedPermissionModes.map((value) => ({
              label: value,
              value,
            })),
          },
        ]
      : [],
    connection: input.connection,
    error: input.error,
    hasOlderItems: input.hasOlderItems,
    interactionMode: session?.interactionMode ?? "acp",
    items,
    metadata: embeddedMetadata(session),
    permissions,
    plan: [],
    sessionId: session?.id ?? "",
    status: input.status,
    terminals: hasTerminal ? input.terminals : [],
    title: session?.title ?? "Agent session",
  };
}

function embeddedMetadata(session: EmbeddedSession | null): AgentSessionMetadata[] {
  if (!session) return [];
  const metadata: AgentSessionMetadata[] = [];
  if (session.runnerId) {
    metadata.push({ id: "runner", label: "Runner", value: session.runnerId });
  }
  if (session.totalCostUsd != null) {
    metadata.push({
      id: "cost",
      label: "Cost",
      value: `$${session.totalCostUsd.toFixed(4)}`,
    });
  }
  return metadata;
}
