"use client";

import type { WorkflowDraft } from "./resource-editor-types";
import { ResourceIdentityFields } from "./ResourceIdentityFields";
import { useResourceReferenceOptions } from "./use-resource-reference-options";
import { WorkflowDefinitionFields } from "./WorkflowDefinitionFields";
import { WorkflowExecutionFields } from "./WorkflowExecutionFields";

interface WorkflowConfigurationPanelProps {
  orgSlug: string;
  draft: WorkflowDraft;
  identityLocked?: boolean;
  onChange: (draft: WorkflowDraft) => void;
}

export function WorkflowConfigurationPanel({
  orgSlug,
  draft,
  identityLocked,
  onChange,
}: WorkflowConfigurationPanelProps) {
  const catalog = useResourceReferenceOptions(orgSlug);
  return (
    <div className="space-y-6">
      <ResourceIdentityFields
        metadata={draft.metadata}
        locked={identityLocked}
        onChange={(metadata) => onChange({ ...draft, metadata })}
      />
      <WorkflowDefinitionFields
        draft={draft}
        catalog={catalog}
        onChange={onChange}
      />
      <WorkflowExecutionFields draft={draft} onChange={onChange} />
    </div>
  );
}
