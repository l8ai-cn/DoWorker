"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { ResourceMetadata } from "./resource-editor-types";

interface ResourceIdentityFieldsProps {
  metadata: ResourceMetadata;
  locked?: boolean;
  onChange: (metadata: ResourceMetadata) => void;
}

export function ResourceIdentityFields({
  metadata,
  locked,
  onChange,
}: ResourceIdentityFieldsProps) {
  const t = useTranslations("resourceEditor");
  const patch = (next: Partial<ResourceMetadata>) => {
    onChange({ ...metadata, ...next });
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
            disabled={locked}
            value={metadata.name}
            onChange={(event) => patch({ name: event.target.value })}
          />
        </FormField>
        <FormField
          label={t("fields.displayName")}
          htmlFor="resource-display-name"
          className="flex-1"
        >
          <Input
            id="resource-display-name"
            value={metadata.displayName ?? ""}
            onChange={(event) => patch({ displayName: event.target.value })}
          />
        </FormField>
      </FormRow>
    </FormFieldGroup>
  );
}
