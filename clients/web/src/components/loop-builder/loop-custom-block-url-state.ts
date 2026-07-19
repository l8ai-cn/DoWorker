import {
  buildCustomBlockDefinition,
  type LoopCustomBlockDefinition,
} from "./loop-custom-block-types";

const HASH_KEY = "loopCustomBlocks";

export function readLoopCustomBlocksFromHash(hash: string): LoopCustomBlockDefinition[] {
  const raw = new URLSearchParams(hash.replace(/^#/, "")).get(HASH_KEY);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    const definitions: LoopCustomBlockDefinition[] = [];
    for (const item of parsed) {
      const result = buildCustomBlockDefinition({
        slug: stringField(item.slug),
        label: stringField(item.label),
        promptTemplate: stringField(item.promptTemplate),
        commandTemplate: stringField(item.commandTemplate),
        acceptTemplate: stringField(item.acceptTemplate),
      }, definitions);
      if (result.definition) definitions.push(result.definition);
    }
    return definitions;
  } catch {
    return [];
  }
}

export function writeLoopCustomBlocksToHash(
  hash: string,
  definitions: readonly LoopCustomBlockDefinition[],
): string {
  const params = new URLSearchParams(hash.replace(/^#/, ""));
  if (definitions.length === 0) {
    params.delete(HASH_KEY);
  } else {
    params.set(HASH_KEY, JSON.stringify(definitions.map((definition) => ({
      acceptTemplate: definition.expansion.acceptTemplate,
      commandTemplate: definition.expansion.commandTemplate,
      label: definition.label,
      promptTemplate: definition.expansion.promptTemplate,
      slug: definition.slug,
    }))));
  }
  const next = params.toString();
  return next ? `#${next}` : "";
}

function stringField(value: unknown): string {
  return typeof value === "string" ? value : "";
}
