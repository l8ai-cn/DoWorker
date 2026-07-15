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
  defaultValue?: unknown;
  required: boolean;
  description?: string;
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

const FIELD_LABEL_KEYS: Record<string, string> = {
  approval_mode: "approvalMode",
  edit_format: "editFormat",
  effort: "effort",
  mcp_enabled: "mcpEnabled",
  model: "model",
  permission_mode: "permissionMode",
  sandbox_mode: "sandboxMode",
};

const FIELD_OPTION_KEYS: Record<string, Record<string, string>> = {
  approval_mode: {
    untrusted: "approvalUntrusted",
    "on-request": "approvalOnRequest",
    never: "approvalNever",
  },
  edit_format: {
    "": "empty",
    whole: "whole",
    diff: "diff",
    udiff: "udiff",
  },
  effort: {
    "": "empty",
    low: "low",
    medium: "medium",
    high: "high",
  },
  permission_mode: {
    default: "permissionDefault",
    plan: "permissionPlan",
    acceptEdits: "permissionAcceptEdits",
    dontAsk: "permissionDontAsk",
    bypassPermissions: "permissionBypassPermissions",
    supervised: "permissionSupervised",
    auto: "permissionAuto",
    bypass: "permissionBypass",
  },
};

export function workerTypeFieldLabel(
  name: string,
  t?: (key: string) => string,
): string {
  const key = FIELD_LABEL_KEYS[name];
  if (key && t) return t(`workerCreate.typeConfig.fields.${key}`);
  return name
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase() + part.slice(1))
    .join(" ");
}

export function workerTypeFieldOptionLabel(
  fieldName: string,
  option: string,
  t?: (key: string) => string,
): string {
  const key = FIELD_OPTION_KEYS[fieldName]?.[option];
  if (key && t) return t(`workerCreate.typeConfig.options.${key}`);
  return option || t?.("workerCreate.typeConfig.useDefault") || option;
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
  return {
    name,
    kind: value.kind,
    options,
    defaultValue: value.default,
    required: value.required === true,
    description: typeof value.description === "string" ? value.description : undefined,
  };
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
