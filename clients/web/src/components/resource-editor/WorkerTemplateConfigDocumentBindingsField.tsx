"use client";

import { useTranslations } from "next-intl";
import type { WorkerConfigDocumentRequirement } from "@/lib/api/facade/podConnect";
import type { WorkerTemplateConfigDocumentBinding } from "./worker-resource-draft-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { environmentBundleCatalogKey } from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { WorkerTemplateReadOnlyReferenceField } from "./WorkerTemplateReadOnlyReferenceField";

interface WorkerTemplateConfigDocumentBindingsFieldProps {
  requirements: WorkerConfigDocumentRequirement[];
  value: WorkerTemplateConfigDocumentBinding[];
  catalog: ResourceReferenceCatalog;
  onChange: (value: WorkerTemplateConfigDocumentBinding[]) => void;
}

export function WorkerTemplateConfigDocumentBindingsField({
  requirements,
  value,
  catalog,
  onChange,
}: WorkerTemplateConfigDocumentBindingsFieldProps) {
  const t = useTranslations("resourceEditor");
  const requirementIds = new Set(
    requirements.map((requirement) => requirement.document_id),
  );
  const unresolvedBindings = value.filter(
    (binding) =>
      binding.configBundleRef.name &&
      !requirementIds.has(binding.documentId),
  );
  const hasBindings = requirements.length > 0 || unresolvedBindings.length > 0;

  return (
    <section className="space-y-3">
      <h4 className="text-sm font-medium">
        {t("fields.configDocumentBindings")}
      </h4>
      {!hasBindings && (
        <p className="text-xs text-muted-foreground">
          {t("references.configurationBundlesUnavailable")}
        </p>
      )}
      {requirements.map((requirement) => {
        const binding = value.find(
          (item) => item.documentId === requirement.document_id,
        );
        return (
          <div
            key={requirement.document_id}
            className="space-y-1 border-l-2 border-border pl-3"
          >
            <ResourceReferenceField
              id={`config-document-${requirement.document_id}`}
              label={requirement.document_id}
              kind="EnvironmentBundle"
              catalogKey={environmentBundleCatalogKey("config")}
              value={binding?.configBundleRef}
              catalog={catalog}
              required={requirement.required}
              onChange={(configBundleRef) => {
                const next = requirements.flatMap((item) => {
                  const reference = item.document_id === requirement.document_id
                    ? configBundleRef
                    : value.find(
                      (current) => current.documentId === item.document_id,
                    )?.configBundleRef;
                  return reference?.name.trim()
                    ? [{
                      documentId: item.document_id,
                      configBundleRef: reference,
                    }]
                    : [];
                });
                onChange(next);
              }}
            />
            <p className="text-xs text-muted-foreground">
              {requirement.format} · {requirement.target_path}
            </p>
          </div>
        );
      })}
      {unresolvedBindings.map((binding) => (
        <div
          key={binding.documentId}
          className="border-l-2 border-border pl-3"
        >
          <WorkerTemplateReadOnlyReferenceField
            id={`config-document-unresolved-${binding.documentId}`}
            label={binding.documentId}
            revisionLabel={t("fields.revision")}
            value={binding.configBundleRef}
          />
        </div>
      ))}
    </section>
  );
}
