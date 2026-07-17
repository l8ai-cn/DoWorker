"use client";

import { useTranslations } from "next-intl";
import { FormField, FormFieldGroup } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  RESOURCE_ID_BINDING_FIELDS,
  type ResourceBindingDraft,
  type ResourceIDBindingKind,
} from "./resource-editor-types";
import { ResourceIdentityFields } from "./ResourceIdentityFields";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceReferenceOptions } from "./use-resource-reference-options";

interface ResourceBindingConfigurationPanelProps {
  orgSlug: string;
  draft: ResourceBindingDraft;
  onChange: (draft: ResourceBindingDraft) => void;
}

export function ResourceBindingConfigurationPanel({
  orgSlug,
  draft,
  onChange,
}: ResourceBindingConfigurationPanelProps) {
  const t = useTranslations("resourceEditor");
  const catalog = useResourceReferenceOptions(orgSlug);
  return (
    <div className="space-y-6">
      <ResourceIdentityFields
        metadata={draft.metadata}
        onChange={(metadata) => onChange({ ...draft, metadata })}
      />
      <FormFieldGroup title={t("sections.binding")}>
        {draft.kind === "ToolBinding" ? (
          <ResourceReferenceField
            id="tool-model-reference"
            label={t("fields.modelRef")}
            kind="ModelBinding"
            value={draft.spec.modelRef}
            catalog={catalog}
            required
            onChange={(modelRef) => {
              if (modelRef) onChange({
                ...draft,
                spec: { modelRef },
              });
            }}
          />
        ) : (
          <BindingIDField
            draft={draft}
            onChange={onChange}
          />
        )}
      </FormFieldGroup>
    </div>
  );
}

function BindingIDField({
  draft,
  onChange,
}: {
  draft: Exclude<ResourceBindingDraft, { kind: "ToolBinding" }>;
  onChange: (draft: ResourceBindingDraft) => void;
}) {
  const t = useTranslations("resourceEditor");
  const field = RESOURCE_ID_BINDING_FIELDS[draft.kind];
  const value = (draft.spec as Record<string, number>)[field];
  return (
    <FormField
      label={bindingIDLabel(t, draft.kind)}
      htmlFor="binding-resource-id"
      required
    >
      <Input
        id="binding-resource-id"
        type="number"
        min={1}
        value={value || ""}
        onChange={(event) => {
          const resourceID = Number(event.target.value);
          onChange({
            ...draft,
            spec: {
              [field]: Number.isSafeInteger(resourceID) ? resourceID : 0,
            },
          } as ResourceBindingDraft);
        }}
      />
    </FormField>
  );
}

function bindingIDLabel(
  t: ReturnType<typeof useTranslations<"resourceEditor">>,
  kind: ResourceIDBindingKind,
): string {
  switch (kind) {
    case "ModelBinding":
      return t("fields.modelResourceId");
    case "Repository":
      return t("fields.repositoryId");
    case "Skill":
      return t("fields.skillId");
    case "KnowledgeBase":
      return t("fields.knowledgeBaseId");
    case "EnvironmentBundle":
      return t("fields.environmentBundleId");
    case "ComputeTarget":
      return t("fields.computeTargetId");
    case "ResourceProfile":
      return t("fields.resourceProfileId");
  }
}
