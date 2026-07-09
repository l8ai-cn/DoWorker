"use client";

import { FileText } from "lucide-react";

interface WorkerAgentInstructionsSectionProps {
  generatedLayer: string;
  t: (key: string) => string;
}

export function WorkerAgentInstructionsSection({
  generatedLayer,
  t,
}: WorkerAgentInstructionsSectionProps) {
  return (
    <div className="space-y-3" data-testid="worker-agent-instructions">
      <div className="flex gap-2">
        <FileText className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
        <div>
          <p className="text-sm font-medium">{t("ide.createPod.agentInstructionsTitle")}</p>
          <p className="text-xs leading-5 text-muted-foreground">
            {t("ide.createPod.agentInstructionsDescription")}
          </p>
        </div>
      </div>

      <pre className="min-h-[120px] overflow-x-auto whitespace-pre-wrap rounded-md border border-border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
        {generatedLayer.trim() ? generatedLayer : t("ide.createPod.agentInstructionsEmptyPreview")}
      </pre>

      <p className="text-xs text-muted-foreground">{t("ide.createPod.agentInstructionsSourceHint")}</p>
    </div>
  );
}
