"use client";

import type { WorkflowData } from "@/stores/workflow";
import { formatDuration } from "@/lib/utils/time";
import { cn } from "@/lib/utils";

interface WorkflowStatsOverviewProps {
  workflow: WorkflowData;
  t: (key: string) => string;
}

const BREAKDOWN_ROWS: { key: "successful_runs" | "failed_runs" | "active_run_count"; labelKey: string; color: string }[] = [
  { key: "successful_runs", labelKey: "workflows.statusCompleted", color: "bg-success" },
  { key: "active_run_count", labelKey: "workflows.statusRunning", color: "bg-warning" },
  { key: "failed_runs", labelKey: "workflows.statusFailed", color: "bg-destructive" },
];

export function WorkflowStatsOverview({ workflow, t }: WorkflowStatsOverviewProps) {
  const successRate = workflow.total_runs > 0
    ? Math.round((workflow.successful_runs / workflow.total_runs) * 100)
    : 0;
  const avg = workflow.avg_duration_sec != null ? Math.round(workflow.avg_duration_sec) : 0;

  return (
    <div className="surface-card p-4">
      <h3 className="mb-3 text-[13px] font-semibold text-foreground">{t("workflows.overviewTitle")}</h3>

      <div className="mb-3 flex items-start justify-between">
        <div className="flex flex-col gap-0.5">
          <span className="text-[11px] text-muted-foreground">{t("workflows.totalRuns")}</span>
          <span className="text-[20px] font-semibold leading-tight text-foreground">{workflow.total_runs}</span>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-[11px] text-muted-foreground">{t("workflows.success")}</span>
          <span className="text-[20px] font-semibold leading-tight text-success">
            {workflow.successful_runs}
            {workflow.total_runs > 0 && <span className="ml-1 text-[13px] font-medium">· {successRate}%</span>}
          </span>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-[11px] text-muted-foreground">{t("workflows.avgDuration")}</span>
          <span className="text-[20px] font-semibold leading-tight text-foreground">
            {avg > 0 ? formatDuration(avg) : "—"}
          </span>
        </div>
      </div>

      <div className="my-3 h-px bg-border" />

      <div className="flex flex-col gap-2">
        {BREAKDOWN_ROWS.map((row) => (
          <div key={row.key} className="flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <span className={cn("h-2 w-2 rounded-full", row.color)} />
              <span className="text-xs text-foreground">{t(row.labelKey)}</span>
            </div>
            <span className="font-mono text-xs text-muted-foreground">{workflow[row.key] ?? 0}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
