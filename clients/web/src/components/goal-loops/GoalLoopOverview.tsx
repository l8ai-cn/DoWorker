import { CircleCheck, CirclePause, Play, SearchCheck } from "lucide-react";
import type { GoalLoopData } from "@/lib/viewModels/goal-loop";

interface GoalLoopOverviewProps {
  loops: GoalLoopData[];
}

const overviewItems = [
  { key: "active", label: "执行中", Icon: Play },
  { key: "verifying", label: "验证中", Icon: SearchCheck },
  { key: "pending", label: "待处理", Icon: CirclePause },
  { key: "finished", label: "已结束", Icon: CircleCheck },
] as const;

export function GoalLoopOverview({ loops }: GoalLoopOverviewProps) {
  const counts = {
    active: loops.filter((loop) => loop.status === "active").length,
    verifying: loops.filter((loop) => loop.status === "verifying").length,
    pending: loops.filter((loop) => loop.status === "draft" || loop.status === "paused").length,
    finished: loops.filter((loop) => ["completed", "failed", "cancelled"].includes(loop.status)).length,
  };

  return (
    <section className="mb-8 rounded-xl bg-surface-raised p-5 shadow-[var(--shadow-soft)] ring-1 ring-border/25">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="font-semibold">执行概览</h2>
          <p className="mt-1 text-sm text-muted-foreground">优先关注正在运行与等待验证的目标。</p>
        </div>
        <p className="rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
          {counts.active + counts.verifying} 个正在处理
        </p>
      </div>
      <dl className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {overviewItems.map(({ key, label, Icon }) => (
          <div key={key} className="rounded-lg bg-surface-muted/60 px-3 py-3">
            <dt className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Icon className="h-3.5 w-3.5" />{label}
            </dt>
            <dd className="mt-1 text-xl font-semibold">{counts[key]}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
