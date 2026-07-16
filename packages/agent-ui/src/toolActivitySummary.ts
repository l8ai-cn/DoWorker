import { englishFileChangeVerb } from "./toolLocalization";

export interface ToolActivitySummary {
  primary?: string;
  result?: string;
}

export function summarizeToolActivity(
  title: string,
  input?: string,
  output?: string,
  detail?: string,
  translateFileChangeVerb: (kind: string) => string = englishFileChangeVerb,
): ToolActivitySummary {
  const value = title.toLowerCase();
  const parsed = parseRecord(input);
  const primary = isFileChange(value)
    ? summarizeFileChange(parsed, translateFileChangeVerb)
    : isCommand(value)
      ? stringField(parsed, "command") ?? compact(input)
      : summarizeGenericInput(parsed) ?? compact(input);

  return {
    primary,
    result: compact(output ?? (!primary ? detail : undefined)),
  };
}

function summarizeFileChange(
  record: Record<string, unknown> | undefined,
  translateFileChangeVerb: (kind: string) => string,
) {
  const changes = Array.isArray(record?.changes) ? record.changes : [];
  const first = recordValue(changes[0]);
  const path = stringField(first, "path");
  if (!path) return undefined;
  const kind = fileChangeKind(first?.kind);
  const remaining = changes.length > 1 ? ` +${changes.length - 1} files` : "";
  return `${translateFileChangeVerb(kind)} ${fileName(path)}${remaining}`;
}

function summarizeGenericInput(record?: Record<string, unknown>) {
  for (const field of ["path", "file", "filename", "url", "query", "prompt", "action"]) {
    const value = stringField(record, field);
    if (value) return compact(value);
  }
  return undefined;
}

function parseRecord(value?: string) {
  if (!value) return undefined;
  try {
    return recordValue(JSON.parse(value));
  } catch {
    return undefined;
  }
}

function recordValue(value: unknown) {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;
}

function stringField(record: Record<string, unknown> | undefined, field: string) {
  const value = record?.[field];
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function fileChangeKind(value: unknown) {
  if (typeof value === "string") return value;
  return stringField(recordValue(value), "type") ?? "";
}

function fileName(path: string) {
  return path.split(/[\\/]/).filter(Boolean).pop() ?? path;
}

function compact(value?: string) {
  const clean = value?.trim();
  if (!clean || clean === "{}" || clean === "null") return undefined;
  const lines = clean.split(/\r?\n/).slice(0, 3).join("\n");
  return lines.length > 240 ? `${lines.slice(0, 237)}...` : lines;
}

function isCommand(value: string) {
  return ["shell", "terminal", "exec", "command"].some((name) => value.includes(name));
}

function isFileChange(value: string) {
  return ["filechange", "file_change", "edit", "write", "patch"].some((name) =>
    value.includes(name),
  );
}
