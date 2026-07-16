"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { ExpertDraft } from "./resource-editor-types";
import { ResourceIdentityFields } from "./ResourceIdentityFields";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceReferenceOptions } from "./use-resource-reference-options";

interface ExpertConfigurationPanelProps {
  orgSlug: string;
  draft: ExpertDraft;
  onChange: (draft: ExpertDraft) => void;
}

export function ExpertConfigurationPanel({
  orgSlug,
  draft,
  onChange,
}: ExpertConfigurationPanelProps) {
  const t = useTranslations("resourceEditor");
  const catalog = useResourceReferenceOptions(orgSlug);
  const setSpec = (patch: Partial<ExpertDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  return (
    <div className="space-y-6">
      <ResourceIdentityFields
        metadata={draft.metadata}
        onChange={(metadata) => onChange({ ...draft, metadata })}
      />
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
          onChange={(promptRef) => setSpec({ promptRef })}
        />
        <FormRow>
          <FormField
            label={t("fields.category")}
            htmlFor="expert-category"
            className="flex-1"
          >
            <Input
              id="expert-category"
              value={draft.spec.category}
              onChange={(event) => setSpec({ category: event.target.value })}
            />
          </FormField>
        </FormRow>
        <FormField
          label={t("fields.description")}
          htmlFor="expert-description"
        >
          <Textarea
            id="expert-description"
            value={draft.spec.description}
            onChange={(event) => setSpec({
              description: event.target.value,
            })}
          />
        </FormField>
        <FormField
          label={t("fields.releaseNotes")}
          htmlFor="expert-release-notes"
        >
          <Textarea
            id="expert-release-notes"
            value={draft.spec.releaseNotes}
            onChange={(event) => setSpec({
              releaseNotes: event.target.value,
            })}
          />
        </FormField>
      </FormFieldGroup>
    </div>
  );
}
