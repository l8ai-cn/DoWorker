import {
  AgentWorkspace,
  createBuiltinContentRenderers,
  createBuiltinToolRenderers,
  type AgentContentRendererRegistration,
  type AgentToolRendererRegistration,
  type AgentWorkspaceLocale,
  type ContentRendererRegistry,
  type ToolRendererRegistry,
} from "@agent-cloud/agent-ui";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";

import { ImageLightboxProvider } from "@/components/ImageLightbox";
import { TooltipProvider } from "@/components/ui/tooltip";
import type { EmbeddedAgentWorkbenchAccess } from "./embeddedAgentWorkbenchAccess";
import {
  createEmbeddedAgentWorkbenchRuntime,
  type EmbeddedAgentWorkbenchRuntime,
} from "./createEmbeddedAgentWorkbenchRuntime";

export function EmbeddedAgentWorkspace({
  access,
  contentRenderers,
  fetch,
  locale = "zh-CN",
  toolRenderers,
}: {
  access: EmbeddedAgentWorkbenchAccess;
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  fetch?: typeof globalThis.fetch;
  locale?: AgentWorkspaceLocale;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { staleTime: 30_000, refetchOnWindowFocus: false },
        },
      }),
  );

  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <ImageLightboxProvider>
          <EmbeddedAgentWorkspaceContent
            access={access}
            contentRenderers={contentRenderers}
            fetch={fetch}
            locale={locale}
            toolRenderers={toolRenderers}
          />
        </ImageLightboxProvider>
      </TooltipProvider>
    </QueryClientProvider>
  );
}

function EmbeddedAgentWorkspaceContent({
  access,
  contentRenderers,
  fetch,
  locale,
  toolRenderers,
}: {
  access: EmbeddedAgentWorkbenchAccess;
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  fetch?: typeof globalThis.fetch;
  locale: AgentWorkspaceLocale;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}) {
  const { baseUrl, getAccessToken, orgSlug, sessionId } = access;
  const [workbench, setWorkbench] = useState<EmbeddedAgentWorkbenchRuntime | null>(null);
  const [error, setError] = useState<string | null>(null);
  const builtinContentRenderers = useMemo(() => createBuiltinContentRenderers(), []);
  const builtinToolRenderers = useMemo(() => createBuiltinToolRenderers(), []);

  useEffect(() => {
    let active = true;
    let opened: EmbeddedAgentWorkbenchRuntime | null = null;
    setWorkbench(null);
    setError(null);
    void createEmbeddedAgentWorkbenchRuntime(
      { baseUrl, getAccessToken, orgSlug, sessionId },
      { fetch },
    ).then(
      (result) => {
        opened = result;
        if (active) setWorkbench(result);
        else result.runtime.close(sessionId);
      },
      () => {
        if (active) {
          setError(
            locale === "zh-CN"
              ? "Worker 会话连接失败，请稍后重试"
              : "Failed to connect to the Worker session. Please try again.",
          );
        }
      },
    );
    return () => {
      active = false;
      opened?.runtime.close(sessionId);
    };
  }, [baseUrl, fetch, getAccessToken, locale, orgSlug, sessionId]);

  if (error) {
    return <WorkspaceState message={error} role="alert" />;
  }
  if (!workbench) {
    return <WorkspaceState message="正在连接 Agent Workspace…" role="status" />;
  }
  return (
    <div className="h-full min-h-0 overflow-hidden">
      <AgentWorkspace
        clientLabel="agent-workspace-iframe"
        contentRenderers={contentRenderers ?? builtinContentRenderers}
        locale={locale}
        presentation="user"
        runtime={workbench.runtime}
        sessionId={sessionId}
        terminalRuntime={workbench.terminalRuntime}
        toolRenderers={toolRenderers ?? builtinToolRenderers}
      />
    </div>
  );
}

function WorkspaceState({ message, role }: { message: string; role: "alert" | "status" }) {
  return (
    <div
      className="flex h-full min-h-0 items-center justify-center px-6 text-center text-sm text-muted-foreground"
      role={role}
    >
      {message}
    </div>
  );
}
