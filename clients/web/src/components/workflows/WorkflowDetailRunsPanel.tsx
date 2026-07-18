"use client";

import { Loader2 } from "lucide-react";
import type { WorkflowData, WorkflowRunData } from "@/stores/workflow";
import { WorkflowStatsOverview } from "@/components/workflows/WorkflowStatsOverview";
import { WorkflowPromptPreview } from "@/components/workflows/WorkflowPromptPreview";
import { WorkflowRunCard } from "@/components/workflows/WorkflowRunCard";

interface WorkflowDetailRunsPanelProps {
  workflow: WorkflowData;
  runs: WorkflowRunData[];
  runsLoading: boolean;
  runsTotalCount: number;
  t: (key: string, params?: Record<string, string | number>) => string;
  onOpenRun: (run: WorkflowRunData) => void;
  onCancelRun: (runId: number) => void;
  onLoadMore: () => void;
  onRevise: () => void;
}

export function WorkflowDetailRunsPanel({
  workflow,
  runs,
  runsLoading,
  runsTotalCount,
  t,
  onOpenRun,
  onCancelRun,
  onLoadMore,
  onRevise,
}: WorkflowDetailRunsPanelProps) {
  const visibleRuns = runs.slice(0, 6);
  const hasMoreToShow = runs.length > 6 || runs.length < runsTotalCount;

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-foreground">{t("workflows.recentRuns")}</h2>
          <span className="text-xs text-muted-foreground">
            {t("workflows.showingLast", { count: visibleRuns.length, total: runsTotalCount })}
          </span>
        </div>

        {runsLoading && runs.length === 0 ? (
          <div className="flex items-center justify-center rounded-md ring-1 ring-border/20 py-10">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          </div>
        ) : runs.length === 0 ? (
          <div className="surface-card bg-muted/30 p-6 text-center text-sm text-muted-foreground">
            {t("workflows.noRuns")}
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {visibleRuns.map((run) => (
              <WorkflowRunCard key={run.id} run={run} t={t} onOpen={onOpenRun} onCancel={onCancelRun} />
            ))}
            {hasMoreToShow && (
              <button
                type="button"
                className="mt-1 self-center text-xs font-medium text-primary hover:underline disabled:opacity-60"
                disabled={runsLoading}
                onClick={onLoadMore}
              >
                {runsLoading ? t("workflows.loadMore") : t("workflows.viewAll")} →
              </button>
            )}
          </div>
        )}
      </section>

      <aside className="flex flex-col gap-4">
        <WorkflowStatsOverview workflow={workflow} t={t} />
        <WorkflowPromptPreview
          workflow={workflow}
          t={t}
          onRevise={onRevise}
        />
      </aside>
    </div>
  );
}
