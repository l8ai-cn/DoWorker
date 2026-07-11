"use client";

import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import type { WorkerCreateStepId } from "../hooks/workerCreateDraft";

export type { WorkerCreateStepId } from "../hooks/workerCreateDraft";

export interface WorkerCreateStepDef {
  id: WorkerCreateStepId;
  label: string;
  summary?: string;
  complete: boolean;
  accessible: boolean;
}

interface WorkerCreateStepperProps {
  steps: WorkerCreateStepDef[];
  current: WorkerCreateStepId;
  onChange: (step: WorkerCreateStepId) => void;
  orientation?: "horizontal" | "vertical";
}

export function WorkerCreateStepper({
  steps,
  current,
  onChange,
  orientation = "horizontal",
}: WorkerCreateStepperProps) {
  if (orientation === "vertical") {
    return (
      <nav aria-label="Worker creation steps" className="w-full">
        <ol className="flex flex-col gap-0">
          {steps.map((step, index) => {
            const isCurrent = step.id === current;
            const showCheck = step.complete && !isCurrent;

            return (
              <li key={step.id} className="relative flex items-start pb-8 last:pb-0">
                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      "absolute left-4 top-8 -ml-px h-[calc(100%-2rem)] w-0.5 rounded-full",
                      step.complete ? "bg-primary/40" : "bg-border",
                    )}
                    aria-hidden
                  />
                )}

                <button
                  type="button"
                  disabled={!step.accessible}
                  aria-current={isCurrent ? "step" : undefined}
                  onClick={() => step.accessible && onChange(step.id)}
                  className={cn(
                    "group relative flex items-start gap-4 text-left transition-colors w-full",
                    step.accessible ? "cursor-pointer" : "cursor-not-allowed opacity-45",
                  )}
                >
                  <span
                    className={cn(
                      "flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold transition-all z-10",
                      isCurrent && "bg-primary text-primary-foreground ring-2 ring-primary/25 ring-offset-2 ring-offset-background",
                      showCheck && "bg-primary/15 text-primary",
                      !isCurrent && !showCheck && step.accessible && "bg-muted text-muted-foreground group-hover:bg-muted/80",
                      !step.accessible && "bg-muted text-muted-foreground",
                    )}
                  >
                    {showCheck ? <Check className="h-4 w-4" /> : step.id}
                  </span>

                  <div className="flex flex-col gap-1 pt-0.5">
                    <span
                      className={cn(
                        "text-sm font-semibold leading-tight",
                        isCurrent ? "text-foreground" : "text-muted-foreground",
                      )}
                    >
                      {step.label}
                    </span>
                    {step.summary && (
                      <span className="text-xs leading-relaxed text-muted-foreground max-w-[14rem]">
                        {step.summary}
                      </span>
                    )}
                  </div>
                </button>
              </li>
            );
          })}
        </ol>
      </nav>
    );
  }

  return (
    <nav aria-label="Worker creation steps" className="mb-6">
      <ol className="flex items-start gap-0">
        {steps.map((step, index) => {
          const isCurrent = step.id === current;
          const showCheck = step.complete && !isCurrent;

          return (
            <li key={step.id} className="flex min-w-0 flex-1 items-start last:flex-none">
              <button
                type="button"
                disabled={!step.accessible}
                aria-current={isCurrent ? "step" : undefined}
                onClick={() => step.accessible && onChange(step.id)}
                className={cn(
                  "group flex min-w-0 flex-1 flex-col items-center gap-1.5 px-1 text-center transition-colors",
                  step.accessible ? "cursor-pointer" : "cursor-not-allowed opacity-45",
                )}
              >
                <span
                  className={cn(
                    "flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold transition-all",
                    isCurrent && "bg-primary text-primary-foreground ring-2 ring-primary/25 ring-offset-2 ring-offset-background",
                    showCheck && "bg-primary/15 text-primary",
                    !isCurrent && !showCheck && step.accessible && "bg-muted text-muted-foreground group-hover:bg-muted/80",
                    !step.accessible && "bg-muted text-muted-foreground",
                  )}
                >
                  {showCheck ? <Check className="h-4 w-4" /> : step.id}
                </span>
                <span
                  className={cn(
                    "text-xs font-medium leading-tight",
                    isCurrent ? "text-foreground" : "text-muted-foreground",
                  )}
                >
                  {step.label}
                </span>
                {step.summary && (
                  <span className="hidden max-w-[8rem] truncate text-[10px] leading-tight text-muted-foreground sm:block">
                    {step.summary}
                  </span>
                )}
              </button>
              {index < steps.length - 1 && (
                <div
                  className={cn(
                    "mx-1 mt-4 h-0.5 min-w-[1rem] flex-1 rounded-full",
                    step.complete ? "bg-primary/40" : "bg-border",
                  )}
                  aria-hidden
                />
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
