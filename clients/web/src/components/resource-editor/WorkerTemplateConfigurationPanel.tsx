"use client";

import type { WorkerTemplateDraft } from "./resource-editor-types";
import { WorkerTemplateBindingsPanel } from "./WorkerTemplateBindingsPanel";
import { WorkerTemplateIdentityPanel } from "./WorkerTemplateIdentityPanel";
import { WorkerTemplateLifecyclePanel } from "./WorkerTemplateLifecyclePanel";
import { WorkerTemplateRuntimePanel } from "./WorkerTemplateRuntimePanel";
import { WorkerTemplateTypeConfigPanel } from "./WorkerTemplateTypeConfigPanel";
import { WorkerTemplateWorkspacePanel } from "./WorkerTemplateWorkspacePanel";
import { useResourceReferenceOptions } from "./use-resource-reference-options";

interface WorkerTemplateConfigurationPanelProps {
  orgSlug: string;
  draft: WorkerTemplateDraft;
  onChange: (draft: WorkerTemplateDraft) => void;
}

export function WorkerTemplateConfigurationPanel(
  { orgSlug, ...props }: WorkerTemplateConfigurationPanelProps,
) {
  const catalog = useResourceReferenceOptions(orgSlug);
  const panelProps = { ...props, catalog };
  return (
    <div className="space-y-6">
      <WorkerTemplateIdentityPanel {...panelProps} />
      <WorkerTemplateBindingsPanel {...panelProps} />
      <WorkerTemplateRuntimePanel {...panelProps} />
      <WorkerTemplateTypeConfigPanel {...panelProps} />
      <WorkerTemplateWorkspacePanel {...panelProps} />
      <WorkerTemplateLifecyclePanel {...panelProps} />
    </div>
  );
}
