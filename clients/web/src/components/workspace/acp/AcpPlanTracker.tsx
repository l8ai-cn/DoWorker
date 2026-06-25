"use client";

import { useAcpSessionField } from "@/stores/acpSession";
import { CheckCircle2, Circle, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

export function AcpPlanTracker({ podKey }: { podKey: string }) {
  const plan = useAcpSessionField(podKey, (s) => s.plan);

  if (plan.length === 0) return null;

  return (
    <div className="border-b px-4 py-2">
      <div className="flex items-center gap-1.5 flex-wrap">
        <span className="text-xs font-medium text-muted-foreground mr-1">Plan</span>
        {plan.map((step, i) => {
          const isCompleted = step.status === "completed";
          const isInProgress = step.status === "in_progress";

          return (
            <span
              key={i}
              className={cn(
                "inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full",
                isCompleted && "bg-success-bg text-success",
                isInProgress && "bg-info-bg text-info",
                !isCompleted && !isInProgress && "bg-muted text-muted-foreground",
              )}
            >
              {isCompleted ? (
                <CheckCircle2 className="h-3 w-3" />
              ) : isInProgress ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <Circle className="h-3 w-3" />
              )}
              {step.title}
            </span>
          );
        })}
      </div>
    </div>
  );
}
