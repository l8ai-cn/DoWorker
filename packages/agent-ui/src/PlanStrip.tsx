import { CheckCircle2, Circle, CircleAlert, Loader2 } from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type { AgentPlanStep } from "./contracts";

export function PlanStrip({ steps }: { steps: AgentPlanStep[] }) {
  const text = useAgentWorkspaceText();
  if (steps.length === 0) return null;
  return (
    <section className="border-b border-border px-3 py-2" aria-label={text.agentPlan}>
      <div className="mb-1.5 text-xs font-medium text-muted-foreground">
        {text.plan}
      </div>
      <ol className="flex flex-wrap gap-x-4 gap-y-1">
        {steps.map((step) => (
          <li className="flex min-w-0 items-center gap-1.5 text-xs" key={step.id}>
            <PlanIcon status={step.status} />
            <span className={step.status === "completed" ? "text-muted-foreground" : ""}>
              {step.title}
            </span>
          </li>
        ))}
      </ol>
    </section>
  );
}

function PlanIcon({ status }: { status: AgentPlanStep["status"] }) {
  if (status === "completed") return <CheckCircle2 className="size-3.5 text-emerald-600" />;
  if (status === "running") return <Loader2 className="size-3.5 animate-spin text-primary" />;
  if (status === "failed") return <CircleAlert className="size-3.5 text-destructive" />;
  return <Circle className="size-3.5 text-muted-foreground" />;
}
