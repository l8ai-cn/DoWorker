import { createRoot } from "react-dom/client";
import type { AgentWorkspaceLocale } from "@do-worker/agent-ui";

import type { EmbedSessionClient } from "./embed-session-api";
import { EmbeddedAgentWorkspace } from "./embed-session/EmbeddedAgentWorkspace";

export interface EmbeddedAgentWorkspaceMount {
  unmount(): void;
}

export function mountEmbeddedAgentWorkspace(
  element: Element,
  input: {
    client: EmbedSessionClient;
    locale?: AgentWorkspaceLocale;
    sessionId: string;
  },
): EmbeddedAgentWorkspaceMount {
  const root = createRoot(element);
  root.render(
    <EmbeddedAgentWorkspace
      client={input.client}
      locale={input.locale}
      sessionId={input.sessionId}
    />,
  );
  return { unmount: () => root.unmount() };
}
