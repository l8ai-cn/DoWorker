"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { WorkerDraft } from "./resource-editor-types";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { ResourceStringMapField } from "./ResourceStringMapField";
import { useResourceReferenceOptions } from "./use-resource-reference-options";

interface WorkerInvocationConfigurationPanelProps {
  orgSlug: string;
  draft: WorkerDraft;
  onChange: (draft: WorkerDraft) => void;
}

export function WorkerInvocationConfigurationPanel({
  orgSlug,
  draft,
  onChange,
}: WorkerInvocationConfigurationPanelProps) {
  const t = useTranslations("resourceEditor");
  const catalog = useResourceReferenceOptions(orgSlug);
  const setMetadata = (patch: Partial<WorkerDraft["metadata"]>) => {
    onChange({ ...draft, metadata: { ...draft.metadata, ...patch } });
  };
  const setSpec = (patch: Partial<WorkerDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  return (
    <div className="space-y-6">
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
      </FormFieldGroup>
      <FormFieldGroup title={t("sections.invocation")}>
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
          onChange={(promptRef) => setSpec({ promptRef })}
        />
        <FormField label={t("fields.alias")} htmlFor="worker-alias">
          <Input
            id="worker-alias"
            value={draft.spec.alias}
            onChange={(event) => setSpec({ alias: event.target.value })}
          />
        </FormField>
        <ResourceStringMapField
          label={t("fields.inputs")}
          value={draft.spec.inputs}
          onChange={(inputs) => setSpec({ inputs })}
        />
      </FormFieldGroup>
    </div>
  );
}
