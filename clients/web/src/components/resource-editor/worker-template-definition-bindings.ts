import type {
  WorkerConfigDocumentRequirement,
  WorkerCredentialRequirement,
} from "@/lib/api/facade/podConnect";
import type {
  ResourceReference,
  WorkerTemplateConfigDocumentBinding,
} from "./resource-editor-types";

export function synchronizeConfigDocumentBindings(
  requirements: WorkerConfigDocumentRequirement[],
  current: WorkerTemplateConfigDocumentBinding[],
): WorkerTemplateConfigDocumentBinding[] {
  return requirements.flatMap((requirement) => {
    const binding = current.find(
      (item) => item.documentId === requirement.document_id,
    );
    return binding?.configBundleRef.name.trim() ? [binding] : [];
  });
}

export function synchronizeCredentialReferences(
  requirements: WorkerCredentialRequirement[],
  current: Record<string, ResourceReference>,
): Record<string, ResourceReference> {
  return Object.fromEntries(requirements
    .filter((requirement) => requirement.source_kind === "credential_bundle")
    .flatMap((requirement) => {
      const reference = current[requirement.target_name];
      return reference ? [[requirement.target_name, reference]] : [];
    }));
}

export function requiredCredentialReferenceFields(
  schema: Record<string, unknown>,
): Set<string> {
  const fields = isRecord(schema.fields) ? schema.fields : {};
  return new Set(Object.entries(fields).flatMap(([name, value]) =>
    isRecord(value) && value.kind === "secret" && value.required === true
      ? [name]
      : []
  ));
}

export function missingRequiredCredentialReference(
  schema: Record<string, unknown>,
  secretRefs: Record<string, ResourceReference>,
): string | null {
  for (const field of requiredCredentialReferenceFields(schema)) {
    if (!secretRefs[field]?.name) return field;
  }
  return null;
}

export function missingCredentialRequirementGroup(
  schema: Record<string, unknown>,
  secretRefs: Record<string, ResourceReference>,
): string[] | null {
  const groups = Array.isArray(schema.credential_requirement_groups)
    ? schema.credential_requirement_groups
    : [];
  for (const group of groups) {
    if (!isRecord(group) || !Array.isArray(group.any_of)) continue;
    const fields = group.any_of.filter(
      (field): field is string => typeof field === "string",
    );
    if (fields.length >= 2 && !fields.some((field) => secretRefs[field]?.name)) {
      return fields;
    }
  }
  return null;
}

export function missingRequiredConfigDocumentReference(
  requirements: WorkerConfigDocumentRequirement[],
  bindings: WorkerTemplateConfigDocumentBinding[],
): string | null {
  for (const requirement of requirements) {
    if (!requirement.required) continue;
    const binding = bindings.find(
      (item) => item.documentId === requirement.document_id,
    );
    if (!binding?.configBundleRef.name.trim()) return requirement.document_id;
  }
  return null;
}

export function credentialBundleTargetNames(
  requirements: WorkerCredentialRequirement[],
): string[] {
  return requirements
    .filter((requirement) => requirement.source_kind === "credential_bundle")
    .map((requirement) => requirement.target_name)
    .sort();
}

export function sameConfigDocumentBindings(
  left: WorkerTemplateConfigDocumentBinding[],
  right: WorkerTemplateConfigDocumentBinding[],
): boolean {
  return JSON.stringify(left) === JSON.stringify(right);
}

export function sameCredentialReferences(
  left: Record<string, ResourceReference>,
  right: Record<string, ResourceReference>,
): boolean {
  return JSON.stringify(left) === JSON.stringify(right);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}
