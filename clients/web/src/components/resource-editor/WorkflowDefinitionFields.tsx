"use client";

import { useTranslations } from "next-intl";
import { FormFieldGroup } from "@/components/ui/form-field";
import type { WorkflowDraft } from "./resource-editor-types";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { ResourceStringMapField } from "./ResourceStringMapField";
import type { ResourceReferenceCatalog } from "./resource-reference-options";

interface WorkflowDefinitionFieldsProps {
  draft: WorkflowDraft;
  catalog: ResourceReferenceCatalog;
  onChange: (draft: WorkflowDraft) => void;
}

export function WorkflowDefinitionFields({
  draft,
  catalog,
  onChange,
}: WorkflowDefinitionFieldsProps) {
  const t = useTranslations("resourceEditor");
  const setSpec = (patch: Partial<WorkflowDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  return (
    <FormFieldGroup title={t("sections.definition")}>
      <ResourceReferenceField
        id="worker-template-reference"
        label={t("fields.workerTemplateRef")}
        kind="WorkerTemplate"
        value={draft.spec.workerTemplateRef}
        catalog={catalog}
        required
        onChange={(workerTemplateRef) => {
          if (workerTemplateRef) setSpec({ workerTemplateRef });
        }}
      />
      <ResourceReferenceField
        id="prompt-reference"
        label={t("fields.promptRef")}
        kind="Prompt"
        value={draft.spec.promptRef}
        catalog={catalog}
        required
        onChange={(promptRef) => {
          if (promptRef) setSpec({ promptRef });
        }}
      />
      <ResourceStringMapField
        label={t("fields.inputs")}
        value={draft.spec.inputs}
        onChange={(inputs) => setSpec({ inputs })}
      />
    </FormFieldGroup>
  );
}
