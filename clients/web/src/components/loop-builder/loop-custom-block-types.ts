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

export function customBlockDefinitionTypeKey(slug: string): string {
  return `loop_custom_${slug.replaceAll("-", "_")}`;
}

export function buildCustomBlockDefinition(
  draft: LoopCustomBlockDraft,
  definitions: readonly LoopCustomBlockDefinition[] = [],
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
    definition: customBlockDefinition(draft, nextCustomBlockVersion(definitions, slug)),
  };
}

export function nextCustomBlockVersion(
  definitions: readonly LoopCustomBlockDefinition[],
  slug: string,
): number {
  return definitions.reduce(
    (version, definition) => definition.slug === slug
      ? Math.max(version, definition.version + 1)
      : version,
    1,
  );
}

export function latestCustomBlockDefinitions(
  definitions: readonly LoopCustomBlockDefinition[],
): LoopCustomBlockDefinition[] {
  const latest = new Map<string, LoopCustomBlockDefinition>();
  for (const definition of definitions) {
    const current = latest.get(definition.slug);
    if (!current || definition.version > current.version) latest.set(definition.slug, definition);
  }
  return [...latest.values()].sort((a, b) => a.slug.localeCompare(b.slug));
}

export function isValidCustomBlockDefinition(
  definition: LoopCustomBlockDefinition,
): boolean {
  if (!Number.isSafeInteger(definition.version) || definition.version < 1) return false;
  const result = buildCustomBlockDefinition(
    {
      slug: definition.slug,
      label: definition.label,
      promptTemplate: definition.expansion.promptTemplate,
      commandTemplate: definition.expansion.commandTemplate,
      acceptTemplate: definition.expansion.acceptTemplate,
    },
  );
  if (!result.definition) return false;
  const expected = customBlockDefinition(
    {
      slug: definition.slug,
      label: definition.label,
      promptTemplate: definition.expansion.promptTemplate,
      commandTemplate: definition.expansion.commandTemplate,
      acceptTemplate: definition.expansion.acceptTemplate,
    },
    definition.version,
  );
  return (
    expected.slug === definition.slug &&
    expected.version === definition.version &&
    expected.label === definition.label &&
    sameStrings(expected.parameters, definition.parameters) &&
    expected.expansion.agentLocalId === definition.expansion.agentLocalId &&
    expected.expansion.verifierLocalId === definition.expansion.verifierLocalId &&
    expected.expansion.promptTemplate === definition.expansion.promptTemplate &&
    expected.expansion.commandTemplate === definition.expansion.commandTemplate &&
    expected.expansion.acceptTemplate === definition.expansion.acceptTemplate
  );
}

function customBlockDefinition(
  draft: LoopCustomBlockDraft,
  version: number,
): LoopCustomBlockDefinition {
  const slug = draft.slug.trim();
  const label = draft.label.trim();
  return {
    slug,
    label,
    version,
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
  };
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
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
