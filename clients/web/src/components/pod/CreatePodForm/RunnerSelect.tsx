"use client";

import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { Runner } from "@/stores/runner";

interface RunnerSelectProps {
  runners: Runner[];
  selectedRunnerId: number | null;
  onSelect: (runnerId: number | null) => void;
  error?: string;
  t: (key: string) => string;
}

export function RunnerSelect({
  runners,
  selectedRunnerId,
  onSelect,
  error,
  t,
}: RunnerSelectProps) {
  const selectedRunner = runners.find((r) => r.id === selectedRunnerId);
  const triggerLabel = selectedRunner
    ? `${selectedRunner.node_id} (${selectedRunner.current_pods}/${selectedRunner.max_concurrent_pods})`
    : t("ide.createPod.runnerAutoSelect");

  return (
    <div>
      <label
        htmlFor="runner-select"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.selectRunner")}
      </label>
      <Select
        value={selectedRunnerId ? String(selectedRunnerId) : ""}
        onValueChange={(value) => onSelect(value ? Number(value) : null)}
      >
        <SelectTrigger
          id="runner-select"
          aria-invalid={!!error}
          aria-describedby={
            error ? "runner-error" : runners.length === 0 ? "runner-help" : undefined
          }
          className={cn(error && "ring-destructive/60 focus:ring-destructive/40")}
        >
          <span className={cn(!selectedRunnerId && "text-muted-foreground")}>
            {triggerLabel}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("ide.createPod.runnerAutoSelect")}</SelectItem>
          {runners.map((runner) => (
            <SelectItem key={runner.id} value={String(runner.id)}>
              {runner.node_id} ({runner.current_pods}/{runner.max_concurrent_pods})
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {error && (
        <p id="runner-error" className="text-xs text-destructive mt-1">
          {error}
        </p>
      )}
      {!error && (
        <p id="runner-help" className="text-xs text-muted-foreground mt-1">
          {runners.length === 0
            ? t("ide.createPod.noRunnersAvailable")
            : t("ide.createPod.runnerHostHint")}
        </p>
      )}
    </div>
  );
}
