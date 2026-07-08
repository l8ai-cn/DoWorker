import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/lib/sessions-api";
import { listLocalProjects } from "@/lib/projects-local";
import { readAuthToken } from "@/lib/auth-store";

export function useLiveProjects() {
  const authed = Boolean(readAuthToken());
  const query = useQuery({
    queryKey: ["live-projects"],
    queryFn: listProjects,
    enabled: authed,
    staleTime: 30_000,
  });

  const local = listLocalProjects();
  const remote = query.data ?? [];
  const names = [...new Set([...local.map((p) => p.name), ...remote])];

  return {
    names,
    local,
    loading: authed && query.isLoading,
    error: query.error instanceof Error ? query.error.message : null,
    refresh: () => query.refetch(),
    isLive: authed,
  };
}
