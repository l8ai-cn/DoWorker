import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useMemo, useState } from "react";

import { type DoWorkerHostConfig, setDoWorkerHostConfig } from "./lib/host";
import { type RoutingApi, basenamedRouting, reactRouterRouting } from "./lib/routing";
import { EmbedProviders } from "./embed-providers";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, refetchOnWindowFocus: false },
  },
});

export interface DoWorkerAppProps extends DoWorkerHostConfig {
  basename?: string;
  routing?: Partial<RoutingApi>;
  isDarkMode?: boolean;
}

export function DoWorkerApp({
  basename,
  routing,
  isDarkMode,
  ...hostConfig
}: DoWorkerAppProps = {}) {
  useState(() => {
    setDoWorkerHostConfig(hostConfig);
    return null;
  });

  const routingApi = useMemo<RoutingApi>(() => {
    const merged: RoutingApi = { ...reactRouterRouting, ...routing };
    return basename ? basenamedRouting(basename, merged) : merged;
  }, [basename, routing]);

  return (
    <QueryClientProvider client={queryClient}>
      <EmbedProviders routing={routingApi} basename={basename} isDarkMode={isDarkMode} />
    </QueryClientProvider>
  );
}
