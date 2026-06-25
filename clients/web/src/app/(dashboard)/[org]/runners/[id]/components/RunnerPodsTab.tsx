"use client";

import { useTranslations } from "next-intl";
import type { RunnerData, RunnerPodData, SandboxStatus } from "@/lib/api";
import { RunnerPodsToolbar } from "./RunnerPodsToolbar";
import { RunnerPodsTable } from "./RunnerPodsTable";
import { RunnerPodsPagination } from "./RunnerPodsPagination";

interface RunnerPodsTabProps {
  runner: RunnerData;
  pods: RunnerPodData[];
  sandboxStatuses: Map<string, SandboxStatus>;
  loadingPods: boolean;
  loadingSandbox: boolean;
  podFilter: string;
  total: number;
  offset: number;
  limit: number;
  onFilterChange: (filter: string) => void;
  onOffsetChange: (offset: number) => void;
  onRefresh: () => void;
  onRefreshSandbox: () => void;
  onResume: (pod: RunnerPodData) => void;
}

export function RunnerPodsTab(props: RunnerPodsTabProps) {
  const t = useTranslations();
  const filterValue = props.podFilter || "all";

  return (
    <div className="space-y-4">
      <RunnerPodsToolbar
        runner={props.runner}
        podFilter={filterValue}
        loadingPods={props.loadingPods}
        loadingSandbox={props.loadingSandbox}
        t={t}
        onFilterChange={(value) => {
          props.onFilterChange(value === "all" ? "" : value);
          props.onOffsetChange(0);
        }}
        onRefresh={props.onRefresh}
        onRefreshSandbox={props.onRefreshSandbox}
      />

      <RunnerPodsTable
        pods={props.pods}
        sandboxStatuses={props.sandboxStatuses}
        t={t}
        onResume={props.onResume}
      />

      <RunnerPodsPagination
        offset={props.offset}
        limit={props.limit}
        total={props.total}
        t={t}
        onOffsetChange={props.onOffsetChange}
      />
    </div>
  );
}
