"use client";

import Link from "next/link";
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
      <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <span>{t("yaml.limits")}</span>
          <Link
            href="/docs/concepts/resource-orchestration#yaml"
            className="font-medium text-primary hover:underline"
          >
            {t("yaml.manual")}
          </Link>
        </div>
        <span>{bytes.toLocaleString()} / 262,144 bytes</span>
      </div>
      {error && (
        <AlertMessage type="error" message={t("yaml.invalid")} />
      )}
    </div>
  );
}
