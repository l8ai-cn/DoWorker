"use client";

import { useMemo } from "react";
import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import type { AgentfileCompletionContext } from "@/lib/codemirror-agentfile";
import { RunnerSelect } from "./RunnerSelect";
import { WorkerImageSelect } from "./WorkerImageSelect";
import { AutomationLevelSelect } from "./AutomationLevelSelect";
import { AgentfileCodeEditor } from "./AgentfileCodeEditor";

interface WorkerSourceModePanelProps {
  form: CreatePodFormState;
  agents: AgentData[];
  runners: RunnerData[];
  repositories: RepositoryData[];
  selectedRunner: RunnerData | null;
  setSelectedRunnerId: (id: number | null) => void;
  configFields: ConfigField[];
  hasOnlineRunners: boolean;
  t: (key: string) => string;
}

export function WorkerSourceModePanel({
  form,
  agents,
  runners,
  repositories,
  selectedRunner,
  setSelectedRunnerId,
  configFields,
  hasOnlineRunners,
  t,
}: WorkerSourceModePanelProps) {
  const completionContext = useMemo<AgentfileCompletionContext>(
    () => ({
      configFields,
      repositories: repositories.map((r) => ({
        slug: r.slug,
        name: r.name,
        default_branch: r.default_branch,
      })),
      envBundles: form.envBundles.map((b) => ({ name: b.name })),
    }),
    [configFields, repositories, form.envBundles],
  );

  return (
    <div className="space-y-5 animate-in fade-in duration-200">
      <p className="text-sm text-muted-foreground">{t("ide.createPod.sourceModeHint")}</p>
      {hasOnlineRunners && (
        <RunnerSelect
          runners={runners}
          selectedRunnerId={selectedRunner?.id ?? null}
          onSelect={setSelectedRunnerId}
          error={form.validationErrors.runner}
          t={t}
        />
      )}
      <WorkerImageSelect
        images={agents}
        selectedImageSlug={form.selectedAgent}
        onSelect={form.setSelectedAgent}
        hasOnlineClusters={hasOnlineRunners}
        error={form.validationErrors.agent}
        t={t}
      />
      <AutomationLevelSelect
        value={form.automationLevel}
        onChange={form.setAutomationLevel}
        t={t}
      />
      <div className="space-y-2">
        <label className="block text-sm font-medium">
          {t("ide.createPod.agentfileLayer")}
        </label>
        <div className="min-h-[280px] overflow-hidden rounded-md border border-border">
          <AgentfileCodeEditor
            value={form.rawLayerText}
            onChange={form.setRawLayerText}
            completionContext={completionContext}
          />
        </div>
        <p className="text-xs text-muted-foreground">
          {t("ide.createPod.agentInstructionsFooter")}
        </p>
      </div>
    </div>
  );
}
