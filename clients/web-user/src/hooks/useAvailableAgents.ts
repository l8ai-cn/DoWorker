import { useQuery } from "@tanstack/react-query";
import { fetchAvailableAgents } from "./availableAgentCatalogComposition";

export type { AvailableAgent, SessionInteractionMode } from "./availableAgentTypes";

interface UseAvailableAgentsOptions {
  enabled?: boolean;
}

export function useAvailableAgents(options: UseAvailableAgentsOptions = {}) {
  return useQuery({
    queryKey: ["available-agents"],
    queryFn: fetchAvailableAgents,
    enabled: options.enabled ?? true,
    staleTime: 30_000,
  });
}
