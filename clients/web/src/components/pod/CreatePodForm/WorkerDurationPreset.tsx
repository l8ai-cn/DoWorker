"use client";

import { cn } from "@/lib/utils";
import type { DestroyPolicy } from "./podLifecycleOptions";

export type WorkerDurationKind = "short" | "long";

interface WorkerDurationPresetProps {
  perpetual: boolean;
  destroyPolicy: DestroyPolicy;
  onSelect: (kind: WorkerDurationKind) => void;
  t: (key: string) => string;
}

export function resolveDurationKind(
  perpetual: boolean,
  destroyPolicy: DestroyPolicy,
): WorkerDurationKind {
  return perpetual || destroyPolicy === "manual" ? "long" : "short";
}

export function WorkerDurationPreset({
  perpetual,
  destroyPolicy,
  onSelect,
  t,
}: WorkerDurationPresetProps) {
  const active = resolveDurationKind(perpetual, destroyPolicy);

  const options: Array<{
    kind: WorkerDurationKind;
    titleKey: string;
    descriptionKey: string;
  }> = [
    {
      kind: "short",
      titleKey: "ide.createPod.durationShortTitle",
      descriptionKey: "ide.createPod.durationShortDescription",
    },
    {
      kind: "long",
      titleKey: "ide.createPod.durationLongTitle",
      descriptionKey: "ide.createPod.durationLongDescription",
    },
  ];

  return (
    <div>
      <label className="mb-2 block text-sm font-medium">
        {t("ide.createPod.durationLabel")}
      </label>
      <div className="grid gap-2 sm:grid-cols-2">
        {options.map((option) => {
          const selected = active === option.kind;
          return (
            <button
              key={option.kind}
              type="button"
              aria-pressed={selected}
              onClick={() => onSelect(option.kind)}
              className={cn(
                "rounded-lg border px-3 py-3 text-left transition-colors",
                selected
                  ? "border-primary bg-primary/5 ring-1 ring-primary/30"
                  : "border-border hover:border-primary/40 hover:bg-muted/40",
              )}
            >
              <span className="block text-sm font-medium">{t(option.titleKey)}</span>
              <span className="mt-1 block text-xs leading-5 text-muted-foreground">
                {t(option.descriptionKey)}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
