"use client";

import { useState } from "react";
import { Play, Save, SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { createGoalLoop, startGoalLoop } from "@/lib/api/facade/goalLoopConnect";
import type { GoalLoopData } from "@/lib/viewModels/goal-loop";
import type { Pod } from "@/stores/pod";
import {
  initialGoalLoopForm,
  optionalNumber,
  type GoalLoopFormState,
  workerLabel,
  workerSnapshotOptions,
} from "./goal-loop-form-state";

interface GoalLoopCreateFormProps {
  orgSlug: string;
  workers: Pod[];
  onCreated: (loop: GoalLoopData) => void;
}

export function GoalLoopCreateForm({ orgSlug, workers, onCreated }: GoalLoopCreateFormProps) {
  const [form, setForm] = useState<GoalLoopFormState>(initialGoalLoopForm);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const availableWorkers = workerSnapshotOptions(workers);
  const criteria = form.criteria.split("\n").map((item) => item.trim()).filter(Boolean);
  const canSubmit = Boolean(
    form.name.trim() &&
    form.workerSnapshotId &&
    form.objective.trim() &&
    criteria.length > 0 &&
    form.verificationCommand.trim(),
  );

  function update<K extends keyof GoalLoopFormState>(key: K, value: GoalLoopFormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function submit(start: boolean) {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      const loop = await createGoalLoop(orgSlug, {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        worker_spec_snapshot_id: Number(form.workerSnapshotId),
        objective: form.objective.trim(),
        acceptance_criteria: criteria,
        verification_command: form.verificationCommand.trim(),
        max_iterations: optionalNumber(form.maxIterations),
        token_budget: optionalNumber(form.tokenBudget),
        timeout_minutes: optionalNumber(form.timeoutMinutes),
        no_progress_limit: optionalNumber(form.noProgressLimit),
        same_error_limit: optionalNumber(form.sameErrorLimit),
        escalation_policy: form.escalationPolicy,
      });
      onCreated(start ? await startGoalLoop(orgSlug, loop.slug) : loop);
      setForm(initialGoalLoopForm);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "创建 Loop 失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="rounded-xl bg-surface-raised p-5 shadow-[var(--shadow-soft)] ring-1 ring-border/30">
      <div className="mb-5 flex items-start gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Play className="h-4 w-4" />
        </div>
        <div>
          <h2 className="font-semibold">创建目标 Loop</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Loop 只执行一次，完成必须由验证命令退出码为 0 判定。
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="goal-loop-name">名称</Label>
          <Input id="goal-loop-name" value={form.name} onChange={(event) => update("name", event.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>执行 Worker</Label>
          <Select value={form.workerSnapshotId} onValueChange={(value) => update("workerSnapshotId", value)}>
            <SelectTrigger><SelectValue placeholder="选择已有 Worker" /></SelectTrigger>
            <SelectContent>
              {availableWorkers.map(({ snapshotId, worker }) => (
                <SelectItem key={snapshotId} value={String(snapshotId)}>
                  {workerLabel(worker)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">Loop 固定使用该 Worker 的不可变配置快照，不复用运行时会话。</p>
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="goal-loop-objective">目标</Label>
          <Textarea id="goal-loop-objective" value={form.objective} onChange={(event) => update("objective", event.target.value)}
            placeholder="例如：修复结算页的税额计算，并让完整测试集通过。" />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="goal-loop-criteria">验收标准</Label>
          <Textarea id="goal-loop-criteria" value={form.criteria} onChange={(event) => update("criteria", event.target.value)}
            placeholder={"每行一条，例如：\n新增税额边界测试\npnpm test 通过"} />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="goal-loop-verification">验证命令</Label>
          <Input id="goal-loop-verification" value={form.verificationCommand}
            onChange={(event) => update("verificationCommand", event.target.value)} placeholder="pnpm test --filter billing" />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="goal-loop-description">说明</Label>
          <Input id="goal-loop-description" value={form.description} onChange={(event) => update("description", event.target.value)}
            placeholder="可选：补充范围、约束或交付背景" />
        </div>
      </div>

      <details className="mt-5 rounded-lg bg-surface-muted/55 p-4">
        <summary className="flex cursor-pointer list-none items-center gap-2 text-sm font-medium">
          <SlidersHorizontal className="h-4 w-4" />
          执行边界
        </summary>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <NumberField label="最大迭代次数" value={form.maxIterations} onChange={(value) => update("maxIterations", value)} />
          <NumberField label="Token 预算（可选）" value={form.tokenBudget} onChange={(value) => update("tokenBudget", value)} />
          <NumberField label="总运行时长（分钟）" value={form.timeoutMinutes} onChange={(value) => update("timeoutMinutes", value)} />
          <NumberField label="无进展阈值" value={form.noProgressLimit} onChange={(value) => update("noProgressLimit", value)} />
          <NumberField label="同错阈值" value={form.sameErrorLimit} onChange={(value) => update("sameErrorLimit", value)} />
          <div className="space-y-2">
            <Label>触发边界后的处理</Label>
            <Select value={form.escalationPolicy} onValueChange={(value) => update("escalationPolicy", value as "pause" | "fail")}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="pause">暂停并等待人工处理</SelectItem>
                <SelectItem value="fail">标记失败并停止</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </details>

      <p className="mt-4 text-xs text-muted-foreground">
        Cron、并发策略、回调地址和历史保留属于 Workflow，不属于目标 Loop。
      </p>
      {error && <p className="mt-3 text-sm text-destructive">{error}</p>}
      <div className="mt-5 flex justify-end gap-2">
        <Button type="button" variant="outline" disabled={!canSubmit || submitting} onClick={() => void submit(false)}>
          <Save className="mr-2 h-4 w-4" />保存草稿
        </Button>
        <Button type="button" disabled={!canSubmit || submitting} loading={submitting} onClick={() => void submit(true)}>
          <Play className="mr-2 h-4 w-4" />创建并启动
        </Button>
      </div>
    </section>
  );
}

function NumberField({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <Input min="1" type="number" value={value} onChange={(event) => onChange(event.target.value)} />
    </div>
  );
}
