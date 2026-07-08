import { Play, Pause, Target } from "lucide-react";
import { cn } from "@/lib/utils";

export type CodexGoalMode = "active" | "paused";

export function NewTaskGoalPanel({
  tokenBudget,
  goalMode,
  successCriteria,
  onTokenBudgetChange,
  onGoalModeChange,
  onSuccessCriteriaChange,
}: {
  tokenBudget: string;
  goalMode: CodexGoalMode;
  successCriteria: string;
  onTokenBudgetChange: (v: string) => void;
  onGoalModeChange: (v: CodexGoalMode) => void;
  onSuccessCriteriaChange: (v: string) => void;
}) {
  return (
    <div className="space-y-3 rounded-2xl border border-primary/30 bg-primary/5 p-3">
      <div className="flex items-center gap-1.5 text-[10.5px] font-semibold uppercase tracking-wider text-primary">
        <Target className="h-3 w-3" /> Codex 目标
      </div>
      <p className="text-[11px] leading-relaxed text-muted-foreground">
        补充运行参数。目标描述请写在上方输入框。
      </p>
      <div>
        <label className="mb-1 block text-[11px] font-medium text-foreground/80">运行模式</label>
        <div className="flex gap-1 rounded-xl bg-surface p-1 ring-1 ring-border/40">
          {(
            [
              { id: "active" as const, label: "立即执行", icon: Play },
              { id: "paused" as const, label: "先暂停", icon: Pause },
            ] as const
          ).map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              type="button"
              onClick={() => onGoalModeChange(id)}
              className={cn(
                "flex flex-1 items-center justify-center gap-1 rounded-lg py-2 text-[11px] font-medium transition",
                goalMode === id
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Icon className="h-3 w-3" />
              {label}
            </button>
          ))}
        </div>
      </div>
      <div>
        <label className="mb-1 block text-[11px] font-medium text-foreground/80">Token 预算（可选）</label>
        <input
          type="number"
          min={1}
          inputMode="numeric"
          value={tokenBudget}
          onChange={(e) => onTokenBudgetChange(e.target.value)}
          placeholder="留空表示不限制"
          className="w-full rounded-lg bg-surface px-2.5 py-2 font-mono text-[12px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
        />
      </div>
      <div>
        <label className="mb-1 block text-[11px] font-medium text-foreground/80">成功标准（可选）</label>
        <textarea
          rows={2}
          value={successCriteria}
          onChange={(e) => onSuccessCriteriaChange(e.target.value)}
          placeholder="例如：主分支 CI 7 天内绿灯率 ≥ 98%"
          className="w-full resize-none rounded-lg bg-surface px-2.5 py-1.5 text-[12px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
        />
      </div>
    </div>
  );
}

export function buildCodexObjective(prompt: string, successCriteria: string): string {
  const base = prompt.trim();
  const criteria = successCriteria.trim();
  if (!criteria) return base;
  return `${base}\n\n成功标准：${criteria}`;
}

export function parseTokenBudget(raw: string): number | null | undefined {
  const trimmed = raw.trim();
  if (!trimmed) return undefined;
  const n = Number(trimmed);
  if (!Number.isFinite(n) || n <= 0) throw new Error("Token 预算须为正整数");
  return Math.floor(n);
}
