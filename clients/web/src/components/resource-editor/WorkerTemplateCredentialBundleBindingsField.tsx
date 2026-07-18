"use client";

import { useTranslations } from "next-intl";
import type { WorkerCredentialRequirement } from "@/lib/api/facade/podConnect";
import type { ResourceReference } from "./resource-editor-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { environmentBundleCatalogKey } from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { WorkerTemplateReadOnlyReferenceField } from "./WorkerTemplateReadOnlyReferenceField";

interface WorkerTemplateCredentialBundleBindingsFieldProps {
  requirements: WorkerCredentialRequirement[];
  requiredFields: Set<string>;
  value: Record<string, ResourceReference>;
  catalog: ResourceReferenceCatalog;
  onChange: (value: Record<string, ResourceReference>) => void;
}

export function WorkerTemplateCredentialBundleBindingsField({
  requirements,
  requiredFields,
  value,
  catalog,
  onChange,
}: WorkerTemplateCredentialBundleBindingsFieldProps) {
  const t = useTranslations("resourceEditor");
  const bundleRequirements = requirements.filter(
    (requirement) => requirement.source_kind === "credential_bundle",
  );
  const requirementTargets = new Set(
    bundleRequirements.map((requirement) => requirement.target_name),
  );
  const unresolvedReferences = Object.entries(value).filter(
    ([targetName, reference]) =>
      reference.name && !requirementTargets.has(targetName),
  );
  if (
    bundleRequirements.length === 0 &&
    unresolvedReferences.length === 0
  ) return null;

  return (
    <section className="space-y-3">
      <h4 className="text-sm font-medium">{t("fields.secretRefs")}</h4>
      <p className="text-sm text-muted-foreground">
        {t("secrets.referenceOnly")}
      </p>
      {bundleRequirements.map((requirement) => (
        <div
          key={requirement.id}
          className="space-y-1 border-l-2 border-border pl-3"
        >
          <ResourceReferenceField
            id={`credential-bundle-${requirement.id}`}
            label={requirement.target_name}
            kind="EnvironmentBundle"
            catalogKey={environmentBundleCatalogKey(
              "credential",
              requirement.target_name,
            )}
            value={value[requirement.target_name]}
            catalog={catalog}
            required={requiredFields.has(requirement.target_name)}
            onChange={(reference) => {
              const next = { ...value };
              if (reference?.name) {
                next[requirement.target_name] = reference;
              } else {
                delete next[requirement.target_name];
              }
              onChange(next);
            }}
          />
          <p className="text-xs text-muted-foreground">
            {requirement.source_ref} -&gt; {requirement.target_kind}
          </p>
        </div>
      ))}
      {unresolvedReferences.map(([targetName, reference]) => (
        <div
          key={targetName}
          className="border-l-2 border-border pl-3"
        >
          <WorkerTemplateReadOnlyReferenceField
            id={`credential-bundle-unresolved-${targetName}`}
            label={targetName}
            revisionLabel={t("fields.revision")}
            value={reference}
          />
        </div>
      ))}
    </section>
  );
}
