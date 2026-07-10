import { useState, useEffect, useMemo } from "react";
import {
  RunnerData,
  AgentData,
  RepositoryData,
} from "@/lib/api";
import { listRunners } from "@/lib/api/facade/runnerConnect";
import { listAgents } from "@/lib/api/facade/agentConnect";
import { agentsSupportedByRunners } from "@/lib/runner-agent-capabilities";
import { sortAgentsForTaskEntry } from "@/lib/task-entry-agent-order";
import { readCurrentOrg } from "@/stores/auth";
import { useRepositories, useRepositoryStore } from "@/stores/repository";

export interface PodCreationData {
  runners: RunnerData[];
  agents: AgentData[];
  repositories: RepositoryData[];
  loading: boolean;
  error: string | null;
  selectedRunner: RunnerData | null;
  setSelectedRunnerId: (id: number | null) => void;
  availableAgents: AgentData[];
}

export function usePodCreationData(enabled: boolean): PodCreationData {
  const [runners, setRunners] = useState<RunnerData[]>([]);
  const [agents, setAgents] = useState<AgentData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRunnerId, setSelectedRunnerId] = useState<number | null>(null);

  const repositories = useRepositories();
  const fetchRepositories = useRepositoryStore((s) => s.fetchRepositories);
  useEffect(() => {
    if (enabled) fetchRepositories();
  }, [enabled, fetchRepositories]);

  useEffect(() => {
    if (!enabled) return;

    let cancelled = false;

    const loadData = async () => {
      setLoading(true);
      setError(null);
      try {
        const orgSlug = readCurrentOrg()?.slug ?? "";
        const [runnersRes, agentsRes] = await Promise.all([
          listRunners(orgSlug),
          listAgents(orgSlug),
        ]);

        if (cancelled) return;

        const allRunners: RunnerData[] = runnersRes.items;
        const onlineRunners = allRunners.filter((r: RunnerData) => r.status === "online");
        setRunners(onlineRunners);

        const seen = new Set<string>();
        const agentList: AgentData[] = [];
        for (const a of [...agentsRes.builtin_agents, ...agentsRes.custom_agents, ...agentsRes.agents]) {
          if (seen.has(a.slug)) continue;
          seen.add(a.slug);
          agentList.push(a);
        }
        setAgents(agentList);
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Failed to load data";
        setError(message);
        console.error("Failed to load pod creation data:", err);
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    loadData();

    return () => {
      cancelled = true;
    };
  }, [enabled]);

  useEffect(() => {
    if (!enabled) {
      setSelectedRunnerId(null);
    }
  }, [enabled]);

  const selectedRunner = useMemo(() => {
    if (!selectedRunnerId) return null;
    return runners.find(r => r.id === selectedRunnerId) || null;
  }, [runners, selectedRunnerId]);

  const availableAgents = useMemo((): AgentData[] => {
    const supported = agentsSupportedByRunners(agents, selectedRunner ? [selectedRunner] : runners);
    return sortAgentsForTaskEntry(supported);
  }, [selectedRunner, runners, agents]);

  return {
    runners,
    agents,
    repositories,
    loading,
    error,
    selectedRunner,
    setSelectedRunnerId,
    availableAgents,
  };
}
