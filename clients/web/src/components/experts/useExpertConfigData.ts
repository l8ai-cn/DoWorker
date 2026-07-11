"use client";

import { useEffect, useState } from "react";
import { listRunners } from "@/lib/api/facade/runnerConnect";
import { readCurrentOrg } from "@/stores/auth";
import { useRepositories, useRepositoryStore } from "@/stores/repository";
import { useWorkflowEnvBundles } from "@/components/workflows/useWorkflowEnvBundles";
import type { RepositoryData, RunnerData } from "@/lib/api";
import type { EnvBundleSummary } from "@/lib/viewModels/envBundleSummary";

export interface ExpertConfigData {
  runners: RunnerData[];
  repositories: RepositoryData[];
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
}

export function useExpertConfigData(open: boolean, agentSlug: string): ExpertConfigData {
  const [runners, setRunners] = useState<RunnerData[]>([]);
  const repositories = useRepositories();
  const fetchRepositories = useRepositoryStore((s) => s.fetchRepositories);
  const { envBundles, loadingBundles } = useWorkflowEnvBundles({
    open,
    agentSlug: agentSlug || null,
  });

  useEffect(() => {
    if (open) fetchRepositories();
  }, [open, fetchRepositories]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    listRunners(readCurrentOrg()?.slug ?? "")
      .then((res) => {
        if (!cancelled) setRunners(res.items.filter((r) => r.status === "online"));
      })
      .catch(() => {
        if (!cancelled) setRunners([]);
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  return { runners, repositories, envBundles, loadingBundles };
}
