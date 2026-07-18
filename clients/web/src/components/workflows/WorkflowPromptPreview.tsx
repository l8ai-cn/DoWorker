"use client";

import type { WorkflowData } from "@/stores/workflow";

interface WorkflowPromptPreviewProps {
  workflow: WorkflowData;
  t: (key: string) => string;
  onRevise?: () => void;
}

export function WorkflowPromptPreview({
  workflow,
  t,
  onRevise,
}: WorkflowPromptPreviewProps) {
  const template = (workflow.prompt_template || "").trim();
  const lines = template ? template.split("\n") : [];

  return (
    <div className="surface-card p-4">
      <div className="mb-2.5 flex items-center justify-between">
        <h3 className="text-[13px] font-semibold text-foreground">{t("workflows.promptTemplate")}</h3>
        {onRevise && (
          <button
            type="button"
            onClick={onRevise}
            className="text-xs font-medium text-primary hover:underline"
          >
            {t("workflows.newRevision")} →
          </button>
        )}
      </div>

      <div className="rounded-md border border-border bg-muted/40 p-3">
        {lines.length === 0 ? (
          <span className="font-mono text-xs text-muted-foreground">—</span>
        ) : (
          <pre className="whitespace-pre-wrap break-words font-mono text-xs leading-5 text-foreground">
            {template}
          </pre>
        )}
      </div>
    </div>
  );
}
