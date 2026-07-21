import type { AgentConnectionStatus } from "@agent-cloud/agent-ui";

import type { EmbedSessionClient } from "@/embed-session-api";
import { buildRelayWebSocketUrl } from "./relayFrameCodec";

export async function createEmbeddedTerminalSocket(
  client: EmbedSessionClient,
  publishStatus: (status: AgentConnectionStatus) => void,
): Promise<WebSocket> {
  if (!client.getRelayConnection) {
    return fail(
      publishStatus,
      new Error("Embedded session does not allow Relay terminal connections"),
    );
  }
  try {
    const relay = await client.getRelayConnection();
    const socket = new WebSocket(buildRelayWebSocketUrl(relay.relayUrl, relay.token));
    socket.binaryType = "arraybuffer";
    return socket;
  } catch (cause) {
    return fail(
      publishStatus,
      cause instanceof Error ? cause : new Error(String(cause)),
    );
  }
}

function fail(
  publishStatus: (status: AgentConnectionStatus) => void,
  error: Error,
): never {
  publishStatus("disconnected");
  throw error;
}
