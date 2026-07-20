import type {
  LoopCustomBlockDefinition,
  LoopCustomBlockReference,
  LoopResolvedCustomBlockDefinition,
} from "./loop-custom-block-types";

const SHA256_HEX = /^[a-f0-9]{64}$/;

export async function resolveLoopCustomBlockDefinition(
  definition: LoopCustomBlockDefinition,
  definitionId: string,
): Promise<LoopResolvedCustomBlockDefinition> {
  if (!definitionId) throw new Error("custom block definition id is required");
  return {
    ...definition,
    definitionId,
    definitionDigest: await customBlockDefinitionDigest(definition),
  };
}

export async function customBlockDefinitionDigest(
  definition: LoopCustomBlockDefinition,
): Promise<string> {
  if (!globalThis.crypto?.subtle) {
    throw new Error("Web Crypto SHA-256 is required for custom block definitions");
  }
  const bytes = new TextEncoder().encode(JSON.stringify({
    slug: definition.slug,
    version: definition.version,
    label: definition.label,
    parameters: definition.parameters,
    expansion: definition.expansion,
  }));
  const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
  return [...new Uint8Array(digest)].map((value) => value.toString(16).padStart(2, "0")).join("");
}

export function customBlockReference(
  definition: LoopResolvedCustomBlockDefinition,
  nodeId: string,
): LoopCustomBlockReference {
  return {
    nodeId,
    definitionId: definition.definitionId,
    slug: definition.slug,
    version: definition.version,
    definitionDigest: definition.definitionDigest,
  };
}

export function referencePinsDefinition(
  reference: LoopCustomBlockReference | undefined,
  definition: LoopResolvedCustomBlockDefinition,
): boolean {
  return Boolean(
    reference &&
    reference.nodeId &&
    reference.definitionId === definition.definitionId &&
    reference.slug === definition.slug &&
    reference.version === definition.version &&
    SHA256_HEX.test(reference.definitionDigest) &&
    reference.definitionDigest === definition.definitionDigest,
  );
}
