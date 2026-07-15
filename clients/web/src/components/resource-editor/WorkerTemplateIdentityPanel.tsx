"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { ResourceReferenceField } from "./ResourceReferenceField";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";

export function WorkerTemplateIdentityPanel({
  draft,
  catalog,
  onChange,
}: WorkerTemplatePanelProps) {
  const t = useTranslations("resourceEditor");
  const setMetadata = (
    patch: Partial<WorkerTemplatePanelProps["draft"]["metadata"]>,
  ) => {
    onChange({ ...draft, metadata: { ...draft.metadata, ...patch } });
  };
  const setSpec = (
    patch: Partial<WorkerTemplatePanelProps["draft"]["spec"]>,
  ) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };

  return (
    <FormFieldGroup title={t("sections.identity")}>
      <FormRow>
        <FormField
          label={t("fields.name")}
          htmlFor="resource-name"
          required
          className="flex-1"
        >
          <Input
            id="resource-name"
            value={draft.metadata.name}
            onChange={(event) => setMetadata({ name: event.target.value })}
          />
        </FormField>
        <FormField
          label={t("fields.displayName")}
          htmlFor="resource-display-name"
          className="flex-1"
        >
          <Input
            id="resource-display-name"
            value={draft.metadata.displayName ?? ""}
            onChange={(event) => setMetadata({
              displayName: event.target.value,
            })}
          />
        </FormField>
      </FormRow>
      <FormRow>
        <FormField
          label={t("fields.workerType")}
          htmlFor="worker-type"
          required
          className="flex-1"
        >
          <Input
            id="worker-type"
            value={draft.spec.workerType}
            onChange={(event) => setSpec({ workerType: event.target.value })}
          />
        </FormField>
        <FormField
          label={t("fields.optionsRevision")}
          htmlFor="options-revision"
          required
          className="flex-1"
        >
          <Input
            id="options-revision"
            value={draft.spec.optionsRevision}
            onChange={(event) => setSpec({
              optionsRevision: event.target.value,
            })}
          />
        </FormField>
      </FormRow>
      <ResourceReferenceField
        id="model-reference"
        label={t("fields.modelRef")}
        kind="ModelBinding"
        value={draft.spec.modelRef}
        catalog={catalog}
        onChange={(modelRef) => setSpec({ modelRef })}
      />
      <FormField label={t("fields.alias")} htmlFor="worker-alias">
        <Input
          id="worker-alias"
          value={draft.spec.metadata.alias}
          onChange={(event) => setSpec({
            metadata: { alias: event.target.value },
          })}
        />
      </FormField>
    </FormFieldGroup>
  );
}
