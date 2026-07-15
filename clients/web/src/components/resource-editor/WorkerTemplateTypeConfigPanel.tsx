"use client";

import { useTranslations } from "next-intl";
import { AlertMessage } from "@/components/ui/alert-message";
import { FormField, FormFieldGroup } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { ResourceReferenceMapField } from "./ResourceReferenceMapField";
import { WorkerTemplateValuesField } from "./WorkerTemplateValuesField";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";

export function WorkerTemplateTypeConfigPanel({
  draft,
  catalog,
  onChange,
}: WorkerTemplatePanelProps) {
  const t = useTranslations("resourceEditor");
  const typeConfig = draft.spec.typeConfig;
  const update = (patch: Partial<typeof typeConfig>) => {
    onChange({
      ...draft,
      spec: {
        ...draft.spec,
        typeConfig: { ...typeConfig, ...patch },
      },
    });
  };
  return (
    <FormFieldGroup
      title={t("sections.typeConfig")}
      className="border-t border-border pt-6"
    >
      <FormField
        label={t("fields.schemaVersion")}
        htmlFor="type-schema-version"
        required
      >
        <Input
          id="type-schema-version"
          type="number"
          min={1}
          value={typeConfig.schemaVersion || ""}
          onChange={(event) => update({
            schemaVersion: positiveInteger(event.target.value),
          })}
        />
      </FormField>
      <WorkerTemplateValuesField
        value={typeConfig.values}
        onChange={(values) => update({ values })}
      />
      <AlertMessage
        type="info"
        message={t("secrets.referenceOnly")}
      />
      <ResourceReferenceMapField
        id="secret-reference"
        label={t("fields.secretRefs")}
        keyLabel={t("fields.configKey")}
        kind="EnvironmentBundle"
        value={typeConfig.secretRefs}
        catalog={catalog}
        onChange={(secretRefs) => update({ secretRefs })}
      />
    </FormFieldGroup>
  );
}

function positiveInteger(value: string): number {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : 0;
}
