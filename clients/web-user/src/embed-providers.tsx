import { useQueryClient } from "@tanstack/react-query";
import { ThemeProvider as NextThemesProvider } from "next-themes";
import { type ReactNode, useCallback, useEffect, useState } from "react";

import App from "./App";
import { ImageLightboxProvider } from "./components/ImageLightbox";
import { TooltipProvider } from "./components/ui/tooltip";
import { RunnerHealthProvider } from "./hooks/RunnerHealthProvider";
import { SessionUpdatesProvider } from "./hooks/SessionUpdatesProvider";
import { CapabilitiesContext } from "./lib/CapabilitiesContext";
import { resolveServerInfo, type ServerInfo } from "./lib/capabilities";
import { EmbeddedProvider } from "./lib/embedded";
import { setEmbedRoot } from "./lib/host";
import { resolveIdentity } from "./lib/identity";
import { type RoutingApi, RoutingProvider } from "./lib/routing";
import { initChatStore } from "./store/chatStore";

const offlineServerInfo: ServerInfo = {
  accounts_enabled: false,
  login_url: null,
  needs_setup: false,
  databricks_features: false,
  managed_sandboxes_enabled: false,
  sandbox_provider: null,
  server_version: null,
  smart_routing_enabled: false,
};

function EmbedCapabilitiesProvider({ children }: { children: ReactNode }) {
  const [info, setInfo] = useState<ServerInfo | "loading">("loading");
  useEffect(() => {
    let active = true;
    void Promise.race([
      resolveServerInfo(),
      new Promise<ServerInfo>((resolve) => setTimeout(() => resolve(offlineServerInfo), 1500)),
    ]).then((resolved) => {
      if (active) setInfo(resolved);
    });
    return () => {
      active = false;
    };
  }, []);
  return <CapabilitiesContext.Provider value={info}>{children}</CapabilitiesContext.Provider>;
}

export function EmbedProviders({
  routing,
  basename,
  isDarkMode,
}: {
  routing: RoutingApi;
  basename?: string;
  isDarkMode?: boolean;
}) {
  const queryClient = useQueryClient();
  useState(() => {
    initChatStore(queryClient);
    void resolveIdentity();
    return null;
  });

  const scopeRef = useCallback((element: HTMLDivElement | null) => {
    setEmbedRoot(element);
  }, []);

  return (
    <div className="do-worker-app" style={{ height: "100%", width: "100%" }}>
      <div
        ref={scopeRef}
        className={isDarkMode ? "dark" : undefined}
        style={{ height: "100%", width: "100%" }}
      >
        <EmbeddedProvider>
          <NextThemesProvider
            attribute="data-omnigent-theme"
            forcedTheme={isDarkMode ? "dark" : "light"}
            enableColorScheme={false}
            disableTransitionOnChange
          >
            <TooltipProvider>
              <ImageLightboxProvider>
                <RoutingProvider value={routing}>
                  <EmbedCapabilitiesProvider>
                    <SessionUpdatesProvider>
                      <RunnerHealthProvider>
                        <App basename={basename} />
                      </RunnerHealthProvider>
                    </SessionUpdatesProvider>
                  </EmbedCapabilitiesProvider>
                </RoutingProvider>
              </ImageLightboxProvider>
            </TooltipProvider>
          </NextThemesProvider>
        </EmbeddedProvider>
      </div>
    </div>
  );
}
