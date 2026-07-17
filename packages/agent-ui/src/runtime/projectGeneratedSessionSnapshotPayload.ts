import type { AgentError } from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  UnsupportedReason,
  type ContentIdentity,
  type StructuredPayload,
  type UnsupportedValue,
} from "@do-worker/proto/agent_workbench/v2/content_pb";

const decoder = new TextDecoder();

export interface DecodedStructuredPayload {
  text: string;
  value: unknown;
}

export function decodeStructuredPayload(
  payload: StructuredPayload | undefined,
): DecodedStructuredPayload | undefined {
  if (!payload) return undefined;
  const text = decoder.decode(payload.data);
  if (!isJsonMediaType(payload.mediaType)) return { text, value: text };
  try {
    const value: unknown = JSON.parse(text);
    return { text: stringifyValue(value), value };
  } catch {
    return { text, value: text };
  }
}

export function formatStructuredPayload(
  payload: StructuredPayload | undefined,
): string | undefined {
  return decodeStructuredPayload(payload)?.text;
}

export function formatUnsupported(value: UnsupportedValue): string {
  const identity = formatContentIdentity(value.identity);
  const reason = unsupportedReason(value.reason);
  const payload = value.payload;
  if (!payload) return `${identity}; reason=${reason}; payload=missing`;
  const decoded = decodeStructuredPayload(payload)?.text ?? "";
  const bytes = Array.from(payload.data, (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join(" ");
  return [
    identity,
    `reason=${reason}`,
    `mediaType=${payload.mediaType || "unknown"}`,
    decoded ? `utf8=${decoded}` : undefined,
    `bytes=${bytes || "(empty)"}`,
  ]
    .filter(Boolean)
    .join("; ");
}

export function formatAgentError(error: AgentError | undefined): string | null {
  if (!error) return null;
  const detail = formatAgentErrorDetail(error);
  return `[${error.code || "agent_error"}] ${detail}`;
}

export function formatAgentErrorDetail(error: AgentError): string {
  const details = formatStructuredPayload(error.details);
  const violations = error.fieldViolations
    .map((violation) => `${violation.field}: ${violation.description}`)
    .join("\n");
  return [error.message, details, violations]
    .filter(Boolean)
    .join("\n");
}

export function stringifyValue(value: unknown): string {
  if (typeof value === "string") return value;
  try {
    return JSON.stringify(
      value,
      (_, item: unknown) =>
        typeof item === "bigint" ? item.toString() : item,
      2,
    );
  } catch {
    return String(value);
  }
}

function formatContentIdentity(identity: ContentIdentity | undefined): string {
  if (!identity) return "identity=missing";
  return [
    `namespace=${identity.namespace || "missing"}`,
    `semanticKey=${identity.semanticKey || "missing"}`,
    `schemaVersion=${identity.schemaVersion || "missing"}`,
    identity.sourceType ? `sourceType=${identity.sourceType}` : undefined,
  ]
    .filter(Boolean)
    .join(", ");
}

function unsupportedReason(reason: UnsupportedReason): string {
  if (reason === UnsupportedReason.UNKNOWN) return "unknown";
  if (reason === UnsupportedReason.UNSUPPORTED) return "unsupported";
  if (reason === UnsupportedReason.INVALID) return "invalid";
  return "unspecified";
}

function isJsonMediaType(mediaType: string): boolean {
  const normalized = mediaType.toLowerCase().split(";")[0]?.trim() ?? "";
  return normalized === "application/json" || normalized.endsWith("+json");
}
