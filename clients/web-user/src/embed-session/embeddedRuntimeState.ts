import type {
  AgentArtifactItem,
  AgentPermissionResolution,
  AgentWorkspaceCapabilities,
  TerminalResource,
} from "@do-worker/agent-ui";

import type { EmbedSessionClient, EmbeddedSession } from "@/embed-session-api";
import type { AnyBlock, UserMessageBlock } from "@/lib/blocks";
import type { SessionStatus } from "@/lib/types";
import type { ActiveResponse } from "@/store/types";
import type {
  EmbeddedAcpConnectionState,
} from "./EmbeddedAcpConfigurationConnection";
import type { EmbeddedAcpConfiguration } from "./embeddedAcpRelayCodec";

export interface EmbeddedRuntimeState {
  activeResponse: ActiveResponse | null;
  blocks: AnyBlock[];
  capabilities: AgentWorkspaceCapabilities;
  connection: "connecting" | "connected" | "reconnecting" | "disconnected";
  configuration: EmbeddedAcpConfiguration;
  configurationConnected: boolean;
  error: string | null;
  hasOlderItems: boolean;
  session: EmbeddedSession | null;
  sessionId: string;
  status: SessionStatus;
  terminals: TerminalResource[];
  workspaceArtifacts: AgentArtifactItem[];
}

export function createRuntimeState(
  client: EmbedSessionClient,
  sessionId = "",
): EmbeddedRuntimeState {
  return {
    activeResponse: null,
    blocks: [],
    capabilities: {
      interrupt: client.interrupt !== undefined,
      resolvePermission: client.resolvePermission !== undefined,
      sendMessage: client.sendMessage !== undefined,
      terminal: client.getTerminals !== undefined,
      updateConfiguration: false,
    },
    connection: "disconnected",
    configuration: {
      model: "",
      permissionMode: "",
      supportedPermissionModes: [],
    },
    configurationConnected: false,
    error: null,
    hasOlderItems: false,
    session: null,
    sessionId,
    status: "idle",
    terminals: [],
    workspaceArtifacts: [],
  };
}

export function applyStreamBlock(
  state: EmbeddedRuntimeState,
  block: AnyBlock,
): EmbeddedRuntimeState {
  const next = { ...state, blocks: mergeBlocks(state.blocks, [block]) };
  if (block.type === "response_start") {
    return {
      ...next,
      activeResponse: {
        responseId: block.responseId,
        state: "streaming",
        error: null,
      },
      status: "running",
    };
  }
  if (block.type !== "response_end") return next;
  const failed = block.status === "failed";
  return {
    ...next,
    activeResponse: {
      responseId: block.ctx.responseId,
      state: failed ? "failed" : block.status === "cancelled" ? "cancelled" : "completed",
      error: failed ? block.response?.error?.message ?? "Agent response failed" : null,
    },
    status: failed ? "failed" : "idle",
  };
}

export function applyEmbeddedAcpConnectionState(
  state: EmbeddedRuntimeState,
  update: EmbeddedAcpConnectionState,
): EmbeddedRuntimeState {
  return {
    ...state,
    configuration:
      update.configuration === undefined
        ? state.configuration
        : { ...state.configuration, ...update.configuration },
    configurationConnected:
      update.connected ?? state.configurationConnected,
  };
}

export function mergeBlocks(first: AnyBlock[], second: AnyBlock[]): AnyBlock[] {
  const seen = new Set<string>();
  return [...first, ...second].filter((block) => {
    const id = block.ctx.itemId;
    if (!id || id.startsWith("embedded-pending-")) return true;
    if (seen.has(id)) return false;
    seen.add(id);
    return true;
  });
}

export function optimisticUserBlock(commandId: string, text: string): UserMessageBlock {
  return {
    type: "user_message",
    ctx: {
      agent: null,
      depth: 0,
      itemId: `embedded-pending-${commandId}`,
      responseId: "",
      timestamp: performance.now(),
      turn: 0,
    },
    content: [{ type: "input_text", text }],
    stableKey: commandId,
  };
}

export function resolvePermissionBlock(
  blocks: AnyBlock[],
  permissionId: string,
  result: AgentPermissionResolution,
): AnyBlock[] {
  const response =
    result.action === "accept"
      ? { action: result.action, content: { answers: result.content.answers } }
      : result;
  return blocks.map((block) =>
    block.type === "elicitation" && block.elicitationId === permissionId
      ? {
          ...block,
          response,
          status: "responded",
        }
      : block,
  );
}
