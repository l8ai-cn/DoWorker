import {
  AlertTriangle,
  CheckCircle2,
  Circle,
  LoaderCircle,
} from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type { UserVideoExecutionStep } from "./userVideoExecutionTrace";

export function UserVideoExecutionTrace({
  steps,
}: {
  steps: readonly UserVideoExecutionStep[];
}) {
  const text = useAgentWorkspaceText().videoExecutionTrace;
  if (steps.length === 0) return null;

  return (
    <section
      aria-label={text.label}
      aria-live="polite"
      className="border-b border-border bg-muted/20 px-4 py-3"
    >
      <div className="mx-auto max-w-4xl">
        <div className="mb-2 text-xs font-medium text-foreground">{text.label}</div>
        <ol className="grid gap-2 sm:grid-cols-4">
          {steps.map((step) => {
            const Icon = stepIcon(step.status);
            return (
              <li className="flex min-w-0 gap-2" key={step.id}>
                <Icon
                  aria-hidden="true"
                  className={`mt-0.5 size-4 shrink-0 ${step.status === "running" ? "animate-spin motion-reduce:animate-none" : ""}`}
                />
                <div className="min-w-0">
                  <div className="text-xs font-medium leading-5">
                    {text.step[step.id]}
                  </div>
                  <div className="text-xs leading-5 text-muted-foreground">
                    {step.detail === "rendering"
                      ? text.rendering(step.progress)
                      : text.detail[step.detail]}
                  </div>
                </div>
              </li>
            );
          })}
        </ol>
      </div>
    </section>
  );
}

function stepIcon(status: UserVideoExecutionStep["status"]) {
  if (status === "completed") return CheckCircle2;
  if (status === "running") return LoaderCircle;
  if (status === "failed") return AlertTriangle;
  return Circle;
}
