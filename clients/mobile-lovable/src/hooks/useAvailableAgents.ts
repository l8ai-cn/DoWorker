import { useQuery } from "@tanstack/react-query";
import { agentPickerOption } from "@/lib/agent-display";
import { listAgents } from "@/lib/sessions-api";
import { readAuthToken } from "@/lib/auth-store";

export function useAvailableAgents() {
  const authed = Boolean(readAuthToken());
  const query = useQuery({
    queryKey: ["available-agents"],
    queryFn: listAgents,
    enabled: authed,
    staleTime: 60_000,
  });

  const agents = (query.data ?? []).map((a) =>
    agentPickerOption(a.id, a.name, a.supportedModes, a.harness),
  );

  return {
    agents,
    loading: authed && query.isLoading,
    error: query.error instanceof Error ? query.error.message : null,
    isLive: authed,
  };
}
