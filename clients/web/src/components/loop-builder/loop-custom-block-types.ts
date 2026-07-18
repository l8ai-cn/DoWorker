import {
  expandBlockTemplate,
  extractBlockTemplateParameters,
  matchBlockTemplate,
} from "@/components/block-programming/block-custom-template-kernel";

const IDENTIFIER = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
export interface LoopCustomBlockExpansion {
  agentLocalId: string;
  verifierLocalId: string;
  promptTemplate: string;
  commandTemplate: string;
  acceptTemplate: string;
}

export interface LoopCustomBlockDefinition {
  slug: string;
  version: number;
  label: string;
  parameters: string[];
  expansion: LoopCustomBlockExpansion;
}

export interface LoopCustomBlockDraft {
  slug: string;
  label: string;
  promptTemplate: string;
  commandTemplate: string;
  acceptTemplate: string;
}

export interface LoopCustomBlockIssue {
  field: keyof LoopCustomBlockDraft;
  code: "identifier" | "required";
}

export function customBlockType(definition: LoopCustomBlockDefinition): string {
  return `loop_custom_${definition.slug.replaceAll("-", "_")}_v${definition.version}`;
}

export function buildCustomBlockDefinition(
  draft: LoopCustomBlockDraft,
): { definition?: LoopCustomBlockDefinition; issues: LoopCustomBlockIssue[] } {
  const issues: LoopCustomBlockIssue[] = [];
  const slug = draft.slug.trim();
  const label = draft.label.trim();
  if (!IDENTIFIER.test(slug) || slug.length < 2 || slug.length > 100) {
    issues.push({ field: "slug", code: "identifier" });
  }
  if (!label) issues.push({ field: "label", code: "required" });
  for (const field of ["promptTemplate", "commandTemplate", "acceptTemplate"] as const) {
    if (!draft[field].trim()) issues.push({ field, code: "required" });
  }
  if (issues.length > 0) return { issues };
  return {
    issues,
    definition: {
      slug,
      label,
      version: 1,
      parameters: extractBlockTemplateParameters([
        draft.promptTemplate,
        draft.commandTemplate,
        draft.acceptTemplate,
      ]),
      expansion: {
        agentLocalId: `${slug}-task`,
        verifierLocalId: `${slug}-check`,
        promptTemplate: draft.promptTemplate,
        commandTemplate: draft.commandTemplate,
        acceptTemplate: draft.acceptTemplate,
      },
    },
  };
}

export function expandTemplate(
  template: string,
  values: Record<string, string>,
): { value: string; missing: string[] } {
  return expandBlockTemplate(template, values);
}

export function matchTemplate(
  template: string,
  value: string,
): Record<string, string> | undefined {
  return matchBlockTemplate(template, value);
}
