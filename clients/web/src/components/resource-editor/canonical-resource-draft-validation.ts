import type {
  ResourceDraft,
  ResourceEditorKind,
} from "./resource-editor-types";
import {
  canonicalResourceSpecShapes,
  type CanonicalObjectShape,
  type CanonicalShape,
} from "./canonical-resource-spec-shapes";

const metadataShape: CanonicalObjectShape = {
  type: "object",
  fields: {
    name: "string",
    namespace: "string",
    displayName: "string",
    labels: { type: "map", value: "string" },
  },
  optional: ["displayName", "labels"],
};

export function assertCanonicalResourceDraft(
  value: unknown,
  expectedKind: ResourceEditorKind,
): ResourceDraft {
  if (
    !isRecord(value) ||
    value.apiVersion !== "agentsmesh.io/v1alpha1" ||
    value.kind !== expectedKind ||
    !isRecord(value.metadata) ||
    !isRecord(value.spec)
  ) {
    throw new Error(`Canonical document must be a ${expectedKind} resource.`);
  }
  assertKnownFields(value, ["apiVersion", "kind", "metadata", "spec"], "document");
  assertNonEmptyString(value.metadata.name, "metadata.name");
  assertNonEmptyString(value.metadata.namespace, "metadata.namespace");
  assertShape(value.metadata, metadataShape, "metadata");
  assertShape(value.spec, canonicalResourceSpecShapes[expectedKind], "spec");
  return value as unknown as ResourceDraft;
}

function assertShape(
  value: unknown,
  shape: CanonicalShape,
  path: string,
): void {
  if (shape === "any") return;
  if (shape === "integer") {
    if (!Number.isSafeInteger(value)) invalidType(path, shape);
    return;
  }
  if (shape === "string" || shape === "boolean") {
    if (typeof value !== shape) invalidType(path, shape);
    return;
  }
  if (shape.type === "array") {
    if (!Array.isArray(value)) invalidType(path, "array");
    value.forEach((item, index) => {
      assertShape(item, shape.item, `${path}[${index}]`);
    });
    return;
  }
  if (shape.type === "map") {
    if (!isRecord(value)) invalidType(path, "map");
    Object.values(value).forEach((entry) => {
      assertShape(entry, shape.value, `${path} entries`);
    });
    return;
  }
  if (!isRecord(value)) invalidType(path, "object");
  assertKnownFields(value, Object.keys(shape.fields), path);
  const optional = new Set(shape.optional);
  for (const [field, fieldShape] of Object.entries(shape.fields)) {
    if (optional.has(field) && !(field in value)) {
      continue;
    }
    assertShape(value[field], fieldShape, `${path}.${field}`);
  }
}

function assertKnownFields(
  value: Record<string, unknown>,
  allowed: readonly string[],
  path: string,
): void {
  const known = new Set(allowed);
  if (Object.keys(value).some((field) => !known.has(field))) {
    throw new Error(`Canonical resource ${path} contains unsupported fields.`);
  }
}

function invalidType(path: string, expected: string): never {
  throw new Error(`Canonical resource ${path} must be ${expected}.`);
}

function assertNonEmptyString(value: unknown, path: string): void {
  if (typeof value !== "string" || value.length === 0) {
    throw new Error(`Canonical resource ${path} must be a non-empty string.`);
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
