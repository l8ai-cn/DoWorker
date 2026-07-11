"use client";

import type { WorkflowData } from "@/stores/workflow";

interface WorkflowPromptPreviewProps {
  workflow: WorkflowData;
  t: (key: string) => string;
  onEdit?: () => void;
}

export function WorkflowPromptPreview({ workflow, t, onEdit }: WorkflowPromptPreviewProps) {
  const template = (workflow.prompt_template || "").trim();
  const lines = template ? template.split("\n") : [];

  return (
    <div className="surface-card p-4">
      <div className="mb-2.5 flex items-center justify-between">
        <h3 className="text-[13px] font-semibold text-foreground">{t("workflows.promptTemplate")}</h3>
        {onEdit && (
          <button
            type="button"
            onClick={onEdit}
            className="text-xs font-medium text-primary hover:underline"
          >
            {t("common.edit")} →
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
