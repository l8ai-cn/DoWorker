import {
  AgentWorkspace,
  type AgentWorkspaceLocale,
} from "@do-worker/agent-ui";
import { useMemo } from "react";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAgentSessionRuntime } from "./EmbeddedAgentSessionRuntime";
import { EmbeddedTerminalRuntime } from "./EmbeddedTerminalRuntime";

export function EmbeddedAgentWorkspace({
  client,
  locale = "zh-CN",
  sessionId,
}: {
  client: EmbedSessionClient;
  locale?: AgentWorkspaceLocale;
  sessionId: string;
}) {
  const runtime = useMemo(
    () => new EmbeddedAgentSessionRuntime(client),
    [client],
  );
  const terminalRuntime = useMemo(
    () => (client.getRelayConnection ? new EmbeddedTerminalRuntime(client) : undefined),
    [client],
  );

  return (
    <div className="h-full min-h-0 overflow-hidden">
      <AgentWorkspace
        clientLabel="agent-workspace-iframe"
        locale={locale}
        runtime={runtime}
        sessionId={sessionId}
        terminalRuntime={terminalRuntime}
      />
    </div>
  );
}
