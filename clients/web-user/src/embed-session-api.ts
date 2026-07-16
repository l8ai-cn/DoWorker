import type {
  AgentArtifactItem,
  AgentPermissionResolution,
  TerminalResource,
} from "@do-worker/agent-ui";
import type { EmbedSessionAccess } from "./embed-context";
import {
  parseEmbeddedItems,
  parseEmbeddedRelayConnection,
  parseEmbeddedSession,
  parseEmbeddedTerminals,
  readEmbeddedJson,
} from "./embed-session-response-parsers";
import type { ConversationItem } from "@/lib/conversationItems";
import { hostFetch } from "@/lib/host";
import type { SessionStatus } from "@/lib/types";
import {
  listEmbeddedWorkspaceArtifacts,
  loadEmbeddedWorkspaceArtifact,
} from "./embed-workspace-artifact-api";

export interface EmbeddedSession {
  agentLabel: string;
  id: string;
  interactionMode: "acp" | "pty";
  podKey: string | null;
  runnerId?: string | null;
  title: string | null;
  totalCostUsd?: number | null;
  status: SessionStatus;
}

export interface EmbeddedItemsPage {
  items: ConversationItem[];
  hasMore: boolean;
}

interface PostEventWire {
  item_id?: unknown;
}

type EmbedFetch = (path: string, init?: RequestInit) => Promise<Response>;

export interface EmbedRelayConnection {
  relayUrl: string;
  token: string;
  podKey: string;
}

export interface EmbedSessionClient {
  getSession(): Promise<EmbeddedSession>;
  getItems(beforeItemId?: string): Promise<EmbeddedItemsPage>;
  openStream(signal: AbortSignal): Promise<Response>;
  loadArtifact?: (fileId: string) => Promise<Blob>;
  listWorkspaceArtifacts?: () => Promise<AgentArtifactItem[]>;
  sendMessage?: (text: string) => Promise<{ itemId: string | null }>;
  interrupt?: () => Promise<void>;
  resolvePermission?: (
    permissionId: string,
    result: AgentPermissionResolution,
  ) => Promise<void>;
  getTerminals?: () => Promise<TerminalResource[]>;
  getRelayConnection?: () => Promise<EmbedRelayConnection>;
  getAcpRelayConnection?: () => Promise<EmbedRelayConnection>;
}

export function createEmbedSessionClient(
  access: EmbedSessionAccess,
  fetcher: EmbedFetch = hostFetch,
): EmbedSessionClient {
  const sessionPath = `/v1/embed/sessions/${encodeURIComponent(access.sessionId)}`;
  const request = (path: string, init?: RequestInit) =>
    fetcher(path, {
      ...init,
      headers: { ...init?.headers, Authorization: `Bearer ${access.accessToken}` },
      cache: "no-store",
    });
  const client: EmbedSessionClient = {
    async getSession() {
      const response = await request(sessionPath);
      return parseEmbeddedSession(await readEmbeddedJson(response));
    },
    async getItems(beforeItemId) {
      const query = new URLSearchParams({ limit: "100", order: "desc" });
      if (beforeItemId) query.set("after", beforeItemId);
      const response = await request(`${sessionPath}/items?${query}`);
      return parseEmbeddedItems(await readEmbeddedJson(response));
    },
    openStream(signal) {
      return request(`${sessionPath}/stream`, {
        headers: { Accept: "text/event-stream" },
        signal,
      });
    },
    async loadArtifact(fileId) {
      if (fileId.startsWith("workspace:")) {
        return loadEmbeddedWorkspaceArtifact(
          request,
          sessionPath,
          fileId.slice("workspace:".length),
        );
      }
      const response = await request(
        `${sessionPath}/resources/files/${encodeURIComponent(fileId)}/content`,
      );
      if (!response.ok) {
        throw new Error(`Embedded session request failed (${response.status})`);
      }
      return response.blob();
    },
    listWorkspaceArtifacts() {
      return listEmbeddedWorkspaceArtifacts(request, sessionPath);
    },
  };
  if (access.capabilities.includes("write")) {
    client.sendMessage = async (text) => {
      const response = await request(`${sessionPath}/events`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          type: "message",
          data: { role: "user", content: [{ type: "input_text", text }] },
        }),
      });
      const body = (await readEmbeddedJson(response)) as PostEventWire;
      return { itemId: typeof body.item_id === "string" ? body.item_id : null };
    };
  }
  if (access.capabilities.includes("control")) {
    client.interrupt = async () => {
      await readEmbeddedJson(
        await request(`${sessionPath}/events`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ type: "interrupt", data: {} }),
        }),
      );
    };
    client.getAcpRelayConnection = async () => {
      const response = await request(`${sessionPath}/acp-relay-connection`);
      return parseEmbeddedRelayConnection(await readEmbeddedJson(response));
    };
  }
  if (access.capabilities.includes("approve")) {
    client.resolvePermission = async (permissionId, result) => {
      await readEmbeddedJson(
        await request(`${sessionPath}/elicitations/${encodeURIComponent(permissionId)}/resolve`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(result),
        }),
      );
    };
  }
  if (access.capabilities.includes("terminal")) {
    client.getTerminals = async () => {
      const response = await request(`${sessionPath}/resources/terminals`);
      return parseEmbeddedTerminals(
        await readEmbeddedJson(response),
        access.capabilities.includes("control"),
      );
    };
  }
  if (access.capabilities.includes("terminal") && access.capabilities.includes("control")) {
    client.getRelayConnection = async () => {
      const response = await request(`${sessionPath}/relay-connection`);
      return parseEmbeddedRelayConnection(await readEmbeddedJson(response));
    };
  }
  return client;
}
