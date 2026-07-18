import {
  AgentSessionConnection,
  AgentSessionRuntimeV2,
  AgentWorkbenchConnectTransport,
} from "@do-worker/agent-ui";

import { createEmbedSessionClient } from "@/embed-session-api";
import { EmbeddedTerminalRuntime } from "./EmbeddedTerminalRuntime";
import type { EmbeddedAgentWorkbenchAccess } from "./embeddedAgentWorkbenchAccess";
import { createEmbeddedArtifactLoader } from "./embeddedArtifactLoader";

export interface EmbeddedAgentWorkbenchRuntime {
  runtime: AgentSessionRuntimeV2;
  terminalRuntime: EmbeddedTerminalRuntime;
}

export async function createEmbeddedAgentWorkbenchRuntime(
  access: EmbeddedAgentWorkbenchAccess,
  options: { fetch?: typeof globalThis.fetch } = {},
): Promise<EmbeddedAgentWorkbenchRuntime> {
  const resources = createEmbedSessionClient(access, options.fetch);
  const metadata = await resources.getSession();
  const transport = new AgentWorkbenchConnectTransport({
    ...access,
    fetch: options.fetch,
  });
  const connection = new AgentSessionConnection(transport);
  const runtime = new AgentSessionRuntimeV2({
    ...metadata,
    connection,
    sessionId: access.sessionId,
    loadArtifact: createEmbeddedArtifactLoader(connection, resources),
  });
  return {
    runtime,
    terminalRuntime: new EmbeddedTerminalRuntime(resources),
  };
}
