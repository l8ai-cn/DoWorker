"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { useWorkerCreateOptions } from "@/components/pod/hooks/useWorkerCreateOptions";
import { AlertMessage } from "@/components/ui/alert-message";
import type { WorkerTemplateDraft } from "./resource-editor-types";
import { WorkerTemplateBindingsPanel } from "./WorkerTemplateBindingsPanel";
import { WorkerTemplateIdentityPanel } from "./WorkerTemplateIdentityPanel";
import { WorkerTemplateLifecyclePanel } from "./WorkerTemplateLifecyclePanel";
import { WorkerTemplateRuntimePanel } from "./WorkerTemplateRuntimePanel";
import { WorkerTemplateTypeConfigPanel } from "./WorkerTemplateTypeConfigPanel";
import { WorkerTemplateWorkspacePanel } from "./WorkerTemplateWorkspacePanel";
import { useResourceReferenceOptions } from "./use-resource-reference-options";
import {
  credentialBundleTargetNames,
  missingCredentialRequirementGroup,
  missingRequiredConfigDocumentReference,
  missingRequiredCredentialReference,
  requiredCredentialReferenceFields,
} from "./worker-template-definition-bindings";
import {
  findUnresolvedWorkerTemplateReference,
  workerTemplateReferenceCatalogError,
  workerTemplateHasReferences,
} from "./worker-template-plan-readiness";
import { workerTemplateRequiresModelBinding } from "./worker-template-runtime-options";

interface WorkerTemplateConfigurationPanelProps {
  orgSlug: string;
  draft: WorkerTemplateDraft;
  onChange: (draft: WorkerTemplateDraft) => void;
  onPlanBlockChange: (reason: string | null) => void;
}

export function WorkerTemplateConfigurationPanel(
  {
    orgSlug,
    onPlanBlockChange,
    ...props
  }: WorkerTemplateConfigurationPanelProps,
) {
  const t = useTranslations("resourceEditor");
  const workerOptions = useWorkerCreateOptions(true, orgSlug, {
    workerTypeSlug: "",
    computeTargetId: 0,
    deploymentMode: "",
  });
  const selectedWorkerType = workerOptions.status === "ready"
    ? workerOptions.data.worker_types.find(
      (option) => option.slug === props.draft.spec.workerType,
    )
    : undefined;
  const planBlockReason = workerOptions.status === "ready"
    ? selectedWorkerType?.selectable
      ? null
      : t("workerOptions.workerTypeUnavailable")
    : workerOptions.status === "error"
      ? workerOptions.error
      : t("workerOptions.loading");
  const modelRequired = workerOptions.status === "ready" &&
    workerTemplateRequiresModelBinding(
      workerOptions.data,
      props.draft.spec.workerType,
    );
  const credentialRequirements =
    selectedWorkerType?.credential_requirements ?? [];
  const requiredCredentialFields = requiredCredentialReferenceFields(
    selectedWorkerType?.config_schema ?? {},
  );
  const missingCredentialField = missingRequiredCredentialReference(
    selectedWorkerType?.config_schema ?? {},
    props.draft.spec.typeConfig.secretRefs,
  );
  const missingCredentialGroup = missingCredentialRequirementGroup(
    selectedWorkerType?.config_schema ?? {},
    props.draft.spec.typeConfig.secretRefs,
  );
  const configDocumentRequirements =
    selectedWorkerType?.config_document_requirements ?? [];
  const missingConfigDocument = missingRequiredConfigDocumentReference(
    configDocumentRequirements,
    props.draft.spec.workspace.configDocumentBindings,
  );
  const catalog = useResourceReferenceOptions(
    orgSlug,
    props.draft.spec.workerType,
    modelRequired ? selectedWorkerType?.model_protocol_adapters ?? [] : [],
    credentialBundleTargetNames(credentialRequirements),
  );
  const unresolvedReference = findUnresolvedWorkerTemplateReference(
    props.draft,
    catalog,
  );
  const hasReferences = workerTemplateHasReferences(props.draft);
  const referenceBlockReason = hasReferences && catalog.loading
    ? t("references.loading")
    : workerTemplateReferenceCatalogError(props.draft, catalog) ??
      (unresolvedReference
        ? t("references.empty", { kind: unresolvedReference.kind })
        : null);
  const effectivePlanBlockReason = planBlockReason ?? referenceBlockReason ??
    (missingCredentialField
      ? t("references.required", { field: missingCredentialField })
      : missingCredentialGroup
        ? t("references.requiredAnyOf", {
          fields: missingCredentialGroup.join(", "),
        })
        : missingConfigDocument
          ? t("references.requiredConfigurationDocument", {
            document: missingConfigDocument,
          })
        : null);
  const panelProps = { ...props, catalog };

  useEffect(() => {
    onPlanBlockChange(effectivePlanBlockReason);
  }, [effectivePlanBlockReason, onPlanBlockChange]);

  return (
    <div className="space-y-6">
      {workerOptions.status === "ready" && effectivePlanBlockReason && (
        <AlertMessage type="warning" message={effectivePlanBlockReason} />
      )}
      <WorkerTemplateIdentityPanel
        {...panelProps}
        modelRequired={modelRequired}
      />
      <WorkerTemplateBindingsPanel {...panelProps} />
      <WorkerTemplateRuntimePanel
        {...panelProps}
        workerOptions={workerOptions}
      />
      <WorkerTemplateTypeConfigPanel
        {...panelProps}
        credentialRequirements={credentialRequirements}
        requiredCredentialFields={requiredCredentialFields}
      />
      <WorkerTemplateWorkspacePanel
        {...panelProps}
        configDocumentRequirements={configDocumentRequirements}
      />
      <WorkerTemplateLifecyclePanel {...panelProps} />
    </div>
  );
}
