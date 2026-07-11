export type WorkerTypeConfigFieldKind =
  | "boolean"
  | "string"
  | "number"
  | "select"
  | "secret";

export interface WorkerTypeConfigField {
  name: string;
  kind: WorkerTypeConfigFieldKind;
  options: string[];
}

export interface WorkerTypeConfigSchema {
  version: number;
  fields: WorkerTypeConfigField[];
}

export function parseWorkerTypeConfigSchema(
  value: Record<string, unknown>,
): WorkerTypeConfigSchema {
  const version = value.version;
  const rawFields = value.fields;
  if (
    typeof version !== "number" ||
    version <= 0 ||
    !isRecord(rawFields)
  ) {
    throw new Error("Worker type config schema is invalid");
  }
  const fields = Object.entries(rawFields)
    .map(([name, definition]) => parseField(name, definition))
    .sort((left, right) => left.name.localeCompare(right.name));
  return { version, fields };
}

export function workerTypeFieldLabel(name: string): string {
  return name
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase() + part.slice(1))
    .join(" ");
}

function parseField(
  name: string,
  value: unknown,
): WorkerTypeConfigField {
  if (!isRecord(value) || !isKind(value.kind)) {
    throw new Error(`Worker type field "${name}" is invalid`);
  }
  const options = Array.isArray(value.options)
    ? value.options.filter((option): option is string => typeof option === "string")
    : [];
  if (value.kind === "select" && options.length === 0) {
    throw new Error(`Worker type field "${name}" has no options`);
  }
  return { name, kind: value.kind, options };
}

function isKind(value: unknown): value is WorkerTypeConfigFieldKind {
  return (
    value === "boolean" ||
    value === "string" ||
    value === "number" ||
    value === "select" ||
    value === "secret"
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}
