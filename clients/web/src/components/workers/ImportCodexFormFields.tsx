"use client";

import { WorkerImageSelect } from "@/components/pod/CreatePodForm/WorkerImageSelect";
import type { AgentData } from "@/lib/api";

interface Props {
  agents: AgentData[];
  sourcePath: string;
  title: string;
  selectedAgentSlug: string | null;
  onSourcePathChange: (value: string) => void;
  onTitleChange: (value: string) => void;
  onAgentChange: (value: string | null) => void;
  t: (key: string) => string;
}

export function ImportCodexFormFields({
  agents,
  sourcePath,
  title,
  selectedAgentSlug,
  onSourcePathChange,
  onTitleChange,
  onAgentChange,
  t,
}: Props) {
  return (
    <>
      <div className="flex flex-col gap-1.5">
        <label
          htmlFor="import-codex-source"
          className="text-xs font-medium text-muted-foreground"
        >
          {t("workers.create.import.sourceLabel")}
        </label>
        <input
          id="import-codex-source"
          data-testid="import-codex-source-input"
          type="text"
          value={sourcePath}
          onChange={(event) => onSourcePathChange(event.target.value)}
          placeholder={t("workers.create.import.sourcePlaceholder")}
          className="rounded-md border border-input bg-background px-3 py-2 font-mono text-xs outline-none transition-colors focus-visible:border-ring"
        />
        <p className="text-[11px] text-muted-foreground">{t("workers.create.import.sourceHint")}</p>
      </div>

      <div className="flex flex-col gap-1.5">
        <label
          htmlFor="import-codex-title"
          className="text-xs font-medium text-muted-foreground"
        >
          {t("workers.create.import.titleLabel")}
        </label>
        <input
          id="import-codex-title"
          data-testid="import-codex-title-input"
          type="text"
          value={title}
          onChange={(event) => onTitleChange(event.target.value)}
          placeholder={t("workers.create.import.titlePlaceholder")}
          className="rounded-md border border-input bg-background px-3 py-2 font-mono text-xs outline-none transition-colors focus-visible:border-ring"
        />
      </div>

      <WorkerImageSelect
        images={agents}
        selectedImageSlug={selectedAgentSlug}
        onSelect={onAgentChange}
        hasOnlineClusters={agents.length > 0}
        t={t}
      />
    </>
  );
}
