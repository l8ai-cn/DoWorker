import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";

import { ImageLightboxProvider } from "@/components/ImageLightbox";
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  clearEmbedContextFromLocation,
  inspectEmbedContextOnce,
  readEmbedContext,
  redeemEmbedContextOnce,
  type EmbedContextBootstrap,
  type EmbedSessionAccess,
} from "@/embed-context";
import { createEmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAgentWorkspace } from "./EmbeddedAgentWorkspace";
import { EMBED_READY_MESSAGE, readAllowedEmbedOpenProof } from "./embedParentHandshake";

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, refetchOnWindowFocus: false } },
});

export function EmbeddedSessionIframe() {
  const [access, setAccess] = useState<EmbedSessionAccess | null>(null);
  const [pendingContext, setPendingContext] = useState<{
    bootstrap: EmbedContextBootstrap;
    context: string;
  } | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    const open = async () => {
      try {
        const context = readEmbedContext(window.location.search);
        if (window.parent === window) {
          throw new Error("Embedded session must be opened in an iframe");
        }
        const bootstrap = await inspectEmbedContextOnce(context);
        if (!active) return;
        clearEmbedContextFromLocation();
        setPendingContext({ bootstrap, context });
      } catch (cause) {
        if (active)
          setError(cause instanceof Error ? cause.message : "Unable to open embedded session");
      }
    };
    void open();
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!pendingContext) return;
    let redemptionStarted = false;
    const onMessage = (event: MessageEvent) => {
      const proof = readAllowedEmbedOpenProof(
        event,
        window.parent,
        pendingContext.bootstrap.parentOrigins,
      );
      if (!proof || redemptionStarted) return;
      redemptionStarted = true;
      void redeemEmbedContextOnce(pendingContext.context, proof).then(
        (nextAccess) => {
          setAccess(nextAccess);
          setPendingContext(null);
        },
        (cause) => {
          setError(cause instanceof Error ? cause.message : "Unable to open embedded session");
        },
      );
    };
    window.addEventListener("message", onMessage);
    pendingContext.bootstrap.parentOrigins.forEach((origin) =>
      window.parent.postMessage({ type: EMBED_READY_MESSAGE, version: 1 }, origin),
    );
    return () => window.removeEventListener("message", onMessage);
  }, [pendingContext]);

  const client = useMemo(() => (access ? createEmbedSessionClient(access) : null), [access]);

  if (error) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-2 px-6 text-center">
        <h1 className="font-medium">无法打开嵌入会话</h1>
        <p className="text-sm text-muted-foreground">
          {error === "embed_context is required"
            ? "此嵌入工作区需要有效的会话上下文。"
            : error}
        </p>
      </div>
    );
  }
  if (!client) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        正在等待嵌入页面建立连接…
      </div>
    );
  }
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <ImageLightboxProvider>
          <EmbeddedAgentWorkspace client={client} sessionId={access!.sessionId} />
        </ImageLightboxProvider>
      </TooltipProvider>
    </QueryClientProvider>
  );
}
