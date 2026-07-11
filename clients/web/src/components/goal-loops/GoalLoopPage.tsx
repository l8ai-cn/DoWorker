"use client";

import { useEffect, useState } from "react";
import { AlertCircle, RefreshCw, Target } from "lucide-react";
import { Button } from "@/components/ui/button";
import { GoalLoopCreateForm } from "@/components/goal-loops/GoalLoopCreateForm";
import { GoalLoopList } from "@/components/goal-loops/GoalLoopList";
import { cancelGoalLoop, listGoalLoops, startGoalLoop, verifyGoalLoop } from "@/lib/api/facade/goalLoopConnect";
import type { GoalLoopData } from "@/lib/viewModels/goal-loop";
import { usePods, usePodStore } from "@/stores/pod";

export function GoalLoopPage({ orgSlug }: { orgSlug: string }) {
  const loopsWorkers = usePods();
  const fetchPods = usePodStore((state) => state.fetchPods);
  const [loops, setLoops] = useState<GoalLoopData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busySlug, setBusySlug] = useState<string>();
  const [reloadVersion, setReloadVersion] = useState(0);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const result = await listGoalLoops(orgSlug);
        if (!cancelled) setLoops(result);
      } catch (cause) {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "加载 Loop 失败");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    void fetchPods();
    return () => { cancelled = true; };
  }, [fetchPods, orgSlug, reloadVersion]);

  function replaceLoop(loop: GoalLoopData) {
    setLoops((current) => {
      const index = current.findIndex((item) => item.slug === loop.slug);
      return index === -1 ? [loop, ...current] : current.map((item) => item.slug === loop.slug ? loop : item);
    });
  }

  async function runAction(loop: GoalLoopData, action: "start" | "verify" | "cancel") {
    setBusySlug(loop.slug);
    setError(null);
    try {
      const result = action === "start"
        ? await startGoalLoop(orgSlug, loop.slug)
        : action === "verify"
          ? await verifyGoalLoop(orgSlug, loop.slug)
          : await cancelGoalLoop(orgSlug, loop.slug);
      replaceLoop(result);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Loop 操作失败");
    } finally {
      setBusySlug(undefined);
    }
  }

  return (
    <div className="min-h-full overflow-y-auto bg-[radial-gradient(circle_at_top_right,color-mix(in_srgb,var(--primary)_12%,transparent),transparent_30rem)]">
      <div className="mx-auto w-full max-w-6xl px-6 py-8">
        <header className="mb-8 flex flex-col gap-3 border-b border-border/45 pb-7 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-[var(--shadow-soft)]">
              <Target className="h-5 w-5" />
            </div>
            <h1 className="text-2xl font-semibold tracking-tight">目标 Loop</h1>
            <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
              用一个明确目标驱动一次自主执行。Loop 不按时间重复运行，只有外部验证命令成功才会完成。
            </p>
          </div>
          <p className="max-w-sm text-xs text-muted-foreground sm:text-right">
            需要定时、事件或 API 重复触发时，请使用 Workflow。
          </p>
        </header>

        <GoalLoopCreateForm orgSlug={orgSlug} workers={loopsWorkers} onCreated={replaceLoop} />

        <section className="mt-8">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-[0.12em] text-muted-foreground">已有 Loop</h2>
            <Button size="sm" variant="ghost" disabled={loading} onClick={() => setReloadVersion((version) => version + 1)}>
              <RefreshCw className="mr-1.5 h-3.5 w-3.5" />刷新
            </Button>
          </div>
          {error && (
            <div className="mb-4 flex items-center gap-2 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              <AlertCircle className="h-4 w-4" />{error}
            </div>
          )}
          {loading ? <p className="py-8 text-center text-sm text-muted-foreground">正在加载 Loop…</p> : (
            <GoalLoopList
              loops={loops}
              busySlug={busySlug}
              onStart={(loop) => runAction(loop, "start")}
              onVerify={(loop) => runAction(loop, "verify")}
              onCancel={(loop) => runAction(loop, "cancel")}
            />
          )}
        </section>
      </div>
    </div>
  );
}
