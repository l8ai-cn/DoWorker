import type { TerminalResource } from "@do-worker/agent-ui";
import type { AgentArtifactTransportContext, AgentAttachmentReference } from "@do-worker/agent-ui";
import type { EmbeddedAgentWorkbenchAccess } from "./embed-session/embeddedAgentWorkbenchAccess";
import {
  parseEmbeddedRelayConnection,
  parseEmbeddedSession,
  parseEmbeddedTerminals,
  readEmbeddedJson,
} from "./embed-session-response-parsers";
import { loadEmbeddedArtifactRepresentation } from "./embed-workspace-artifact-api";
import { uploadEmbeddedAttachment } from "./embed-attachment-api";

export interface EmbeddedSession {
  agentLabel: string;
  interactionMode: "acp" | "pty";
  title: string;
}

export interface EmbedRelayConnection {
  relayUrl: string;
  token: string;
  podKey: string;
}

export interface EmbedSessionClient {
  getSession(): Promise<EmbeddedSession>;
  uploadAttachment(file: File): Promise<AgentAttachmentReference>;
  loadDownload(downloadUrl: string): Promise<Blob>;
  loadResource(resourceId: string, context: AgentArtifactTransportContext): Promise<Blob>;
  getTerminals(): Promise<TerminalResource[]>;
  getRelayConnection(): Promise<EmbedRelayConnection>;
}

export function createEmbedSessionClient(
  access: EmbeddedAgentWorkbenchAccess,
  fetcher: typeof globalThis.fetch = globalThis.fetch,
): EmbedSessionClient {
  const baseUrl = requiredBaseUrl(access.baseUrl);
  const baseOrigin = new URL(baseUrl).origin;
  const sessionPath = `/v1/embed/sessions/${encodeURIComponent(access.sessionId)}`;
  const request = async (path: string, init?: RequestInit) => {
    const token = (await access.getAccessToken()).trim();
    if (!token) throw new Error("agent_workbench_access_token_missing");
    return fetcher(new URL(path, baseUrl), {
      ...init,
      headers: { ...init?.headers, Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
  };
  const client: EmbedSessionClient = {
    async getSession() {
      const response = await request(sessionPath);
      return parseEmbeddedSession(await readEmbeddedJson(response));
    },
    uploadAttachment(file) {
      return uploadEmbeddedAttachment(request, sessionPath, file);
    },
    async loadResource(resourceId, context) {
      return loadEmbeddedArtifactRepresentation(request, sessionPath, {
        artifactId: context.artifactId,
        digest: context.representation.digest ?? "",
        representationId: context.representationId ?? "",
        resourceId,
        revision: context.descriptor.revision,
      });
    },
    async loadDownload(downloadUrl) {
      const target = new URL(downloadUrl, baseUrl);
      if (target.origin !== baseOrigin) {
        throw new Error("agent_workbench_download_origin_mismatch");
      }
      const response = await request(target.toString());
      await requireOk(response);
      return response.blob();
    },
    async getTerminals() {
      const response = await request(`${sessionPath}/resources/terminals`);
      return parseEmbeddedTerminals(await readEmbeddedJson(response), true);
    },
    async getRelayConnection() {
      const response = await request(`${sessionPath}/relay-connection`);
      return parseEmbeddedRelayConnection(await readEmbeddedJson(response));
    },
  };
  return client;
}

function requiredBaseUrl(value: string): string {
  const normalized = value.trim();
  if (!normalized) throw new Error("agent_workbench_base_url_missing");
  return `${normalized.replace(/\/+$/, "")}/`;
}

async function requireOk(response: Response): Promise<void> {
  if (!response.ok) {
    throw new Error(`Embedded session request failed (${response.status})`);
  }
}
