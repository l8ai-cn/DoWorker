"use client";

import { CircleCheck, CirclePause, CircleX, Play, SearchCheck, Square } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { GoalLoopData } from "@/lib/viewModels/goal-loop";

interface GoalLoopListProps {
  loops: GoalLoopData[];
  busySlug?: string;
  onStart: (loop: GoalLoopData) => Promise<void>;
  onVerify: (loop: GoalLoopData) => Promise<void>;
  onCancel: (loop: GoalLoopData) => Promise<void>;
}

const statusStyle: Record<GoalLoopData["status"], string> = {
  draft: "bg-surface-muted text-muted-foreground",
  active: "bg-primary/10 text-primary",
  paused: "bg-warning/10 text-warning",
  verifying: "bg-primary/10 text-primary",
  completed: "bg-success/10 text-success",
  failed: "bg-destructive/10 text-destructive",
  cancelled: "bg-surface-muted text-muted-foreground",
};

const statusLabel: Record<GoalLoopData["status"], string> = {
  draft: "草稿",
  active: "执行中",
  paused: "已暂停",
  verifying: "验证中",
  completed: "已完成",
  failed: "失败",
  cancelled: "已取消",
};

function statusIcon(status: GoalLoopData["status"]) {
  if (status === "completed") return CircleCheck;
  if (status === "failed") return CircleX;
  if (status === "paused") return CirclePause;
  if (status === "verifying") return SearchCheck;
  return Square;
}

export function GoalLoopList({ loops, busySlug, onStart, onVerify, onCancel }: GoalLoopListProps) {
  if (loops.length === 0) {
    return <div className="rounded-xl border border-dashed border-border/70 p-8 text-center text-sm text-muted-foreground">尚未创建目标 Loop。</div>;
  }

  return (
    <div className="space-y-3">
      {loops.map((loop) => {
        const Icon = statusIcon(loop.status);
        const busy = busySlug === loop.slug;
        return (
          <article key={loop.id} className="rounded-xl bg-surface-raised p-5 shadow-[var(--shadow-soft)] ring-1 ring-border/25">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="font-semibold">{loop.name}</h2>
                  <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${statusStyle[loop.status]}`}>
                    <Icon className="h-3.5 w-3.5" />{statusLabel[loop.status]}
                  </span>
                </div>
                <p className="mt-2 whitespace-pre-wrap text-sm text-muted-foreground">{loop.objective}</p>
                <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span>Worker 快照 #{loop.worker_spec_snapshot_id}</span>
                  <span>最多 {loop.max_iterations} 次迭代</span>
                  <span>总时长 {loop.timeout_minutes} 分钟</span>
                  {loop.pod_key && <span>Pod: {loop.pod_key}</span>}
                </div>
              </div>
              <div className="flex shrink-0 gap-2">
                {(loop.status === "draft" || loop.status === "paused") && (
                  <Button size="sm" disabled={busy} loading={busy} onClick={() => void onStart(loop)}>
                    <Play className="mr-1.5 h-3.5 w-3.5" />启动
                  </Button>
                )}
                {loop.status === "active" && (
                  <Button size="sm" variant="outline" disabled={busy} loading={busy} onClick={() => void onVerify(loop)}>
                    <SearchCheck className="mr-1.5 h-3.5 w-3.5" />验证
                  </Button>
                )}
                {(loop.status === "active" || loop.status === "verifying") && (
                  <Button size="sm" variant="destructive" disabled={busy} onClick={() => void onCancel(loop)}>
                    取消
                  </Button>
                )}
              </div>
            </div>
            {loop.verification_error && <p className="mt-4 rounded-md bg-destructive/8 px-3 py-2 text-xs text-destructive">{loop.verification_error}</p>}
            {loop.verification_output && (
              <pre className="mt-4 max-h-40 overflow-auto rounded-md bg-surface-muted p-3 text-xs text-muted-foreground">{loop.verification_output}</pre>
            )}
          </article>
        );
      })}
    </div>
  );
}
