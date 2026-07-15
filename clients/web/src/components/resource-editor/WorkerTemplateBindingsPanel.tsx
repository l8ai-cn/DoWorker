"use client";

import { useTranslations } from "next-intl";
import { FormFieldGroup } from "@/components/ui/form-field";
import { ResourceReferenceMapField } from "./ResourceReferenceMapField";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";

export function WorkerTemplateBindingsPanel({
  draft,
  catalog,
  onChange,
}: WorkerTemplatePanelProps) {
  const t = useTranslations("resourceEditor");
  return (
    <FormFieldGroup
      title={t("sections.bindings")}
      className="border-t border-border pt-6"
    >
      <ResourceReferenceMapField
        id="tool-reference"
        label={t("fields.toolRefs")}
        keyLabel={t("fields.toolRole")}
        kind="ToolBinding"
        value={draft.spec.toolRefs}
        catalog={catalog}
        onChange={(toolRefs) => onChange({
          ...draft,
          spec: { ...draft.spec, toolRefs },
        })}
      />
    </FormFieldGroup>
  );
}
