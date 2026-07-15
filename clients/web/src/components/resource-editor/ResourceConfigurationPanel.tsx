"use client";

import type { ResourceDraft } from "./resource-editor-types";
import { isResourceBindingDraft } from "./resource-editor-types";
import { ExpertConfigurationPanel } from "./ExpertConfigurationPanel";
import { GoalLoopConfigurationPanel } from "./GoalLoopConfigurationPanel";
import { PromptConfigurationPanel } from "./PromptConfigurationPanel";
import { ResourceBindingConfigurationPanel } from "./ResourceBindingConfigurationPanel";
import { WorkerInvocationConfigurationPanel } from "./WorkerInvocationConfigurationPanel";
import { WorkerTemplateConfigurationPanel } from "./WorkerTemplateConfigurationPanel";
import { WorkflowConfigurationPanel } from "./WorkflowConfigurationPanel";

interface ResourceConfigurationPanelProps {
  orgSlug: string;
  draft: ResourceDraft;
  onChange: (draft: ResourceDraft) => void;
}

export function ResourceConfigurationPanel({
  orgSlug,
  draft,
  onChange,
}: ResourceConfigurationPanelProps) {
  if (isResourceBindingDraft(draft)) {
    return <ResourceBindingConfigurationPanel
      orgSlug={orgSlug}
      draft={draft}
      onChange={onChange}
    />;
  }
  switch (draft.kind) {
    case "Worker":
      return <WorkerInvocationConfigurationPanel
        orgSlug={orgSlug}
        draft={draft}
        onChange={onChange}
      />;
    case "WorkerTemplate":
      return <WorkerTemplateConfigurationPanel
        orgSlug={orgSlug}
        draft={draft}
        onChange={onChange}
      />;
    case "Prompt":
      return <PromptConfigurationPanel draft={draft} onChange={onChange} />;
    case "Expert":
      return <ExpertConfigurationPanel
        orgSlug={orgSlug}
        draft={draft}
        onChange={onChange}
      />;
    case "Workflow":
      return <WorkflowConfigurationPanel
        orgSlug={orgSlug}
        draft={draft}
        onChange={onChange}
      />;
    case "GoalLoop":
      return <GoalLoopConfigurationPanel
        orgSlug={orgSlug}
        draft={draft}
        onChange={onChange}
      />;
  }
}
