import { createRoot } from "react-dom/client";
import type {
  AgentContentRendererRegistration,
  AgentToolRendererRegistration,
  AgentWorkspaceLocale,
  ContentRendererRegistry,
  ToolRendererRegistry,
} from "@agent-cloud/agent-ui";

import { EmbeddedAgentWorkspace } from "./embed-session/EmbeddedAgentWorkspace";
import type { EmbeddedAgentWorkbenchAccess } from "./embed-session/embeddedAgentWorkbenchAccess";

export interface EmbeddedAgentWorkspaceMount {
  unmount(): void;
}

export function mountEmbeddedAgentWorkspace(
  element: Element,
  input: {
    access: EmbeddedAgentWorkbenchAccess;
    contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
    fetch?: typeof globalThis.fetch;
    locale?: AgentWorkspaceLocale;
    toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
  },
): EmbeddedAgentWorkspaceMount {
  const hadScopeClass = element.classList.contains("agent-cloud-app");
  element.classList.add("agent-cloud-app");
  const root = createRoot(element);
  root.render(
    <EmbeddedAgentWorkspace
      access={input.access}
      contentRenderers={input.contentRenderers}
      fetch={input.fetch}
      locale={input.locale}
      toolRenderers={input.toolRenderers}
    />,
  );
  return {
    unmount: () => {
      root.unmount();
      if (!hadScopeClass) element.classList.remove("agent-cloud-app");
    },
  };
}
