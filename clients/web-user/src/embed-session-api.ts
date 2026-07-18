import type { TerminalResource } from "@do-worker/agent-ui";
import type { EmbeddedAgentWorkbenchAccess } from "./embed-session/embeddedAgentWorkbenchAccess";
import {
  parseEmbeddedRelayConnection,
  parseEmbeddedSession,
  parseEmbeddedTerminals,
  readEmbeddedJson,
} from "./embed-session-response-parsers";
import {
  loadEmbeddedWorkspaceArtifact,
  type EmbeddedArtifactIdentity,
} from "./embed-workspace-artifact-api";

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
  loadDownload(downloadUrl: string): Promise<Blob>;
  loadResource(
    resourceId: string,
    identity?: EmbeddedArtifactIdentity,
  ): Promise<Blob>;
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
    async loadResource(resourceId, identity) {
      if (resourceId.startsWith("workspace:")) {
        if (!identity) throw new Error("artifact_identity_missing");
        return loadEmbeddedWorkspaceArtifact(
          request,
          sessionPath,
          resourceId.slice("workspace:".length),
          identity,
        );
      }
      const response = await request(
        `${sessionPath}/resources/files/${encodeURIComponent(resourceId)}/content`,
      );
      await requireOk(response);
      return response.blob();
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
