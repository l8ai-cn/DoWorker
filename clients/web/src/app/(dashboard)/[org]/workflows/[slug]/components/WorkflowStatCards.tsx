import { Hash, CheckCircle2, XCircle, Timer } from "lucide-react";
import type { WorkflowData } from "@/stores/workflow";
import { formatDuration } from "@/lib/utils/time";
import { StatCard } from "./StatCard";

interface WorkflowStatCardsProps {
  workflow: WorkflowData;
  t: (key: string) => string;
}

export function WorkflowStatCards({ workflow, t }: WorkflowStatCardsProps) {
  const successRate =
    workflow.total_runs > 0
      ? Math.round((workflow.successful_runs / workflow.total_runs) * 100)
      : 0;
  const avgDuration = workflow.avg_duration_sec != null ? Math.round(workflow.avg_duration_sec) : 0;

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-8">
      <StatCard icon={Hash} label={t("workflows.totalRuns")} value={workflow.total_runs.toString()} />
      <StatCard
        icon={CheckCircle2}
        iconColor="text-success"
        label={t("workflows.success")}
        value={workflow.successful_runs.toString()}
        suffix={workflow.total_runs > 0 ? `${successRate}%` : undefined}
      />
      <StatCard icon={XCircle} iconColor="text-danger" label={t("workflows.failed")} value={workflow.failed_runs.toString()} />
      <StatCard icon={Timer} label={t("workflows.avgDuration")} value={avgDuration > 0 ? formatDuration(avgDuration) : "-"} />
    </div>
  );
}
