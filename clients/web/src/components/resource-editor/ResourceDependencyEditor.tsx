"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { FormField } from "@/components/ui/form-field";
import type { ResourceEditorKind } from "./resource-editor-types";
import { ResourceEditorShell } from "./ResourceEditorShell";

const DEPENDENCY_KINDS = [
  "Prompt",
  "ModelBinding",
  "ToolBinding",
  "Repository",
  "Skill",
  "KnowledgeBase",
  "EnvironmentBundle",
  "ComputeTarget",
  "ResourceProfile",
] as const satisfies readonly ResourceEditorKind[];

export function ResourceDependencyEditor({ orgSlug }: { orgSlug: string }) {
  const t = useTranslations("resourceEditor");
  const [kind, setKind] = useState<(typeof DEPENDENCY_KINDS)[number]>("Prompt");
  return (
    <div className="space-y-6">
      <FormField
        label={t("fields.resourceKind")}
        htmlFor="resource-kind"
        className="max-w-sm"
      >
        <select
          id="resource-kind"
          className="h-9 w-full rounded-md bg-surface-raised px-3 text-sm ring-1 ring-border/35 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35"
          value={kind}
          onChange={(event) => setKind(
            event.target.value as (typeof DEPENDENCY_KINDS)[number],
          )}
        >
          {DEPENDENCY_KINDS.map((option) => (
            <option key={option} value={option}>{option}</option>
          ))}
        </select>
      </FormField>
      <ResourceEditorShell key={kind} orgSlug={orgSlug} kind={kind} />
    </div>
  );
}
