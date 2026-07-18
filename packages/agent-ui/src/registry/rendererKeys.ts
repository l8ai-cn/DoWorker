export interface ToolRendererKey {
  namespace: string;
  semanticKey: string;
  schemaVersion: string;
}

export interface ContentRendererKey {
  blockKind: string;
  mediaType?: string;
  role?: string;
  schemaVersion: string;
}

export function toolRendererKeyId(key: ToolRendererKey): string {
  return JSON.stringify([
    requiredKeyField("namespace", key.namespace),
    requiredKeyField("semanticKey", key.semanticKey),
    requiredKeyField("schemaVersion", key.schemaVersion),
  ]);
}

export function contentRendererKeyId(key: ContentRendererKey): string {
  return JSON.stringify([
    requiredKeyField("blockKind", key.blockKind),
    optionalKeyField("mediaType", key.mediaType),
    optionalKeyField("role", key.role),
    requiredKeyField("schemaVersion", key.schemaVersion),
  ]);
}

function requiredKeyField(name: string, value: unknown): string {
  if (typeof value !== "string" || value.length === 0) {
    throw new Error(`renderer_key_invalid: field=${name}`);
  }
  return value;
}

function optionalKeyField(
  name: string,
  value: unknown,
): string | undefined {
  if (value === undefined) {
    return undefined;
  }
  return requiredKeyField(name, value);
}
