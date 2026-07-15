"use client";

import { useTranslations } from "next-intl";
import { FormField, FormFieldGroup } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";
import type { PromptDraft } from "./resource-editor-types";
import { PromptVariablesField } from "./PromptVariablesField";
import { ResourceIdentityFields } from "./ResourceIdentityFields";

interface PromptConfigurationPanelProps {
  draft: PromptDraft;
  onChange: (draft: PromptDraft) => void;
}

export function PromptConfigurationPanel({
  draft,
  onChange,
}: PromptConfigurationPanelProps) {
  const t = useTranslations("resourceEditor");
  const setSpec = (patch: Partial<PromptDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  return (
    <div className="space-y-6">
      <ResourceIdentityFields
        metadata={draft.metadata}
        onChange={(metadata) => onChange({ ...draft, metadata })}
      />
      <FormFieldGroup title={t("sections.prompt")}>
        <FormField
          label={t("fields.promptContent")}
          htmlFor="prompt-content"
          required
        >
          <Textarea
            id="prompt-content"
            className="min-h-48 font-mono text-xs"
            value={draft.spec.content}
            onChange={(event) => setSpec({ content: event.target.value })}
          />
        </FormField>
        <PromptVariablesField
          value={draft.spec.variables}
          onChange={(variables) => setSpec({ variables })}
        />
      </FormFieldGroup>
    </div>
  );
}
