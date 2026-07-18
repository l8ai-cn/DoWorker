import { loadWorkerCreateOptions } from "@/components/pod/hooks/useWorkerCreateOptions";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { environmentBundleCatalogKey } from "./resource-reference-options";
import { loadResourceReferenceCatalog } from "./resource-reference-catalog-loader";
import type {
  ResourceReference,
  WorkerTemplateDraft,
} from "./resource-editor-types";
import {
  credentialBundleTargetNames,
  missingCredentialRequirementGroup,
  missingRequiredConfigDocumentReference,
  missingRequiredCredentialReference,
} from "./worker-template-definition-bindings";

interface CatalogReference {
  kind: string;
  key: string;
  reference?: ResourceReference;
}

export async function assertWorkerTemplatePlanReady(
  orgSlug: string,
  draft: WorkerTemplateDraft,
): Promise<void> {
  const options = await loadWorkerCreateOptions(orgSlug, {
    workerTypeSlug: draft.spec.workerType,
    computeTargetId: 0,
    deploymentMode: draft.spec.runtime.deploymentMode,
  });
  const selectedType = options.worker_types.find(
    (option) => option.slug === draft.spec.workerType,
  );
  if (!selectedType?.selectable) {
    throw new Error("The selected Worker type is unavailable.");
  }
  if (draft.spec.optionsRevision !== options.revision) {
    throw new Error("Worker options changed. Review the current catalog.");
  }
  const image = options.runtime_images.find(
    (option) =>
      option.id === draft.spec.runtime.runtimeImageId &&
      option.worker_type_slugs.includes(draft.spec.workerType),
  );
  if (!image?.selectable) {
    throw new Error("The selected runtime image is unavailable.");
  }
  const missingCredentialField = missingRequiredCredentialReference(
    selectedType.config_schema,
    draft.spec.typeConfig.secretRefs,
  );
  if (missingCredentialField) {
    throw new Error(
      `Credential reference "${missingCredentialField}" is required.`,
    );
  }
  const missingCredentialGroup = missingCredentialRequirementGroup(
    selectedType.config_schema,
    draft.spec.typeConfig.secretRefs,
  );
  if (missingCredentialGroup) {
    throw new Error(
      `At least one credential reference is required: ${missingCredentialGroup.join(", ")}.`,
    );
  }
  const missingConfigDocument = missingRequiredConfigDocumentReference(
    selectedType.config_document_requirements,
    draft.spec.workspace.configDocumentBindings,
  );
  if (missingConfigDocument) {
    throw new Error(
      `Configuration document "${missingConfigDocument}" requires an EnvironmentBundle reference.`,
    );
  }
  const catalog = await loadResourceReferenceCatalog(
    orgSlug,
    draft.spec.workerType,
    credentialBundleTargetNames(selectedType.credential_requirements),
  );
  const catalogError = workerTemplateReferenceCatalogError(draft, catalog);
  if (catalogError) throw new Error(catalogError);
  const unresolved = findUnresolvedWorkerTemplateReference(draft, catalog);
  if (unresolved) {
    throw new Error(
      `The referenced ${unresolved.kind} resource "${unresolved.name}" is unavailable.`,
    );
  }
}

export function findUnresolvedWorkerTemplateReference(
  draft: WorkerTemplateDraft,
  catalog: ResourceReferenceCatalog,
): { kind: string; name: string } | null {
  for (const item of workerTemplateReferences(draft)) {
    const name = item.reference?.name;
    if (!name) continue;
    const options = catalog.byKind[item.key] ?? [];
    if (!options.some((option) => option.name === name)) {
      return { kind: item.kind, name };
    }
  }
  return null;
}

export function workerTemplateHasReferences(
  draft: WorkerTemplateDraft,
): boolean {
  return workerTemplateReferences(draft).some(
    (item) => Boolean(item.reference?.name),
  );
}

export function workerTemplateReferenceCatalogError(
  draft: WorkerTemplateDraft,
  catalog: ResourceReferenceCatalog,
): string | null {
  for (const item of workerTemplateReferences(draft)) {
    if (!item.reference?.name) continue;
    const error = catalog.errorsByKind[item.key] ?? catalog.error;
    if (error) return error;
  }
  return null;
}

function workerTemplateReferences(
  draft: WorkerTemplateDraft,
): CatalogReference[] {
  const spec = draft.spec;
  return [
    reference("ModelBinding", spec.modelRef),
    ...Object.values(spec.toolRefs).map((value) =>
      reference("ToolBinding", value)
    ),
    reference("ComputeTarget", spec.runtime.computeTargetRef),
    reference("ResourceProfile", spec.runtime.resourceProfileRef),
    reference("Repository", spec.workspace.repositoryRef),
    ...spec.workspace.skillRefs.map((value) => reference("Skill", value)),
    ...spec.workspace.knowledgeMounts.map((mount) =>
      reference("KnowledgeBase", mount.ref)
    ),
    ...spec.workspace.environmentBundleRefs.map((value) => ({
      kind: "EnvironmentBundle",
      key: environmentBundleCatalogKey("runtime"),
      reference: value,
    })),
    ...spec.workspace.configDocumentBindings.map((binding) => ({
      kind: "EnvironmentBundle",
      key: environmentBundleCatalogKey("config"),
      reference: binding.configBundleRef,
    })),
    ...Object.entries(spec.typeConfig.secretRefs).map(
      ([targetName, value]) => ({
        kind: "EnvironmentBundle",
        key: environmentBundleCatalogKey("credential", targetName),
        reference: value,
      }),
    ),
  ];
}

function reference(
  kind: string,
  value?: ResourceReference,
): CatalogReference {
  return { kind, key: kind, reference: value };
}
