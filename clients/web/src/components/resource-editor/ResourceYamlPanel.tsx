"use client";

import { useTranslations } from "next-intl";
import { AlertMessage } from "@/components/ui/alert-message";
import { FormField } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";
import type { ResourceEditorKind } from "./resource-editor-types";

interface ResourceYamlPanelProps {
  kind: ResourceEditorKind;
  value: string;
  error: string | null;
  onChange: (value: string) => void;
}

export function ResourceYamlPanel({
  kind,
  value,
  error,
  onChange,
}: ResourceYamlPanelProps) {
  const t = useTranslations("resourceEditor");
  const bytes = new TextEncoder().encode(value).byteLength;
  return (
    <div className="space-y-3">
      <FormField label={t("yaml.label", { kind })} htmlFor="resource-yaml">
        <Textarea
          id="resource-yaml"
          data-testid="resource-yaml-editor"
          spellCheck={false}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          className="min-h-[32rem] resize-y overflow-x-auto whitespace-pre font-mono text-xs leading-5"
          error={error ?? undefined}
        />
      </FormField>
      <div className="flex justify-end text-xs text-muted-foreground">
        {bytes.toLocaleString()} / 262,144 bytes
      </div>
      {error && (
        <AlertMessage type="error" message={t("yaml.invalid")} />
      )}
    </div>
  );
}
