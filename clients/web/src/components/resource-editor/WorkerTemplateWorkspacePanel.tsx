"use client";

import { useTranslations } from "next-intl";
import { FormField, FormFieldGroup } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ResourceReferenceListField } from "./ResourceReferenceListField";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { WorkerTemplateConfigDocumentBindingsField } from "./WorkerTemplateConfigDocumentBindingsField";
import { WorkerTemplateKnowledgeMountsField } from "./WorkerTemplateKnowledgeMountsField";
import { environmentBundleCatalogKey } from "./resource-reference-options";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";
import type { WorkerConfigDocumentRequirement } from "@/lib/api/facade/podConnect";

export function WorkerTemplateWorkspacePanel({
  draft,
  catalog,
  configDocumentRequirements,
  onChange,
}: WorkerTemplatePanelProps & {
  configDocumentRequirements: WorkerConfigDocumentRequirement[];
}) {
  const t = useTranslations("resourceEditor");
  const setWorkspace = (
    patch: Partial<WorkerTemplatePanelProps["draft"]["spec"]["workspace"]>,
  ) => {
    onChange({
      ...draft,
      spec: {
        ...draft.spec,
        workspace: { ...draft.spec.workspace, ...patch },
      },
    });
  };
  return (
    <FormFieldGroup
      title={t("sections.workspace")}
      className="border-t border-border pt-6"
    >
      <ResourceReferenceField
        id="repository-reference"
        label={t("fields.repositoryRef")}
        kind="Repository"
        value={draft.spec.workspace.repositoryRef}
        catalog={catalog}
        onChange={(repositoryRef) => setWorkspace({ repositoryRef })}
      />
      <FormField label={t("fields.branch")} htmlFor="workspace-branch">
        <Input
          id="workspace-branch"
          value={draft.spec.workspace.branch}
          onChange={(event) => setWorkspace({ branch: event.target.value })}
        />
      </FormField>
      <ResourceReferenceListField
        id="skill-reference"
        label={t("fields.skillRefs")}
        kind="Skill"
        value={draft.spec.workspace.skillRefs}
        catalog={catalog}
        onChange={(skillRefs) => setWorkspace({ skillRefs })}
      />
      <WorkerTemplateKnowledgeMountsField
        value={draft.spec.workspace.knowledgeMounts}
        catalog={catalog}
        onChange={(knowledgeMounts) => setWorkspace({ knowledgeMounts })}
      />
      <ResourceReferenceListField
        id="environment-bundle-reference"
        label={t("fields.environmentBundleRefs")}
        kind="EnvironmentBundle"
        catalogKey={environmentBundleCatalogKey("runtime")}
        value={draft.spec.workspace.environmentBundleRefs}
        catalog={catalog}
        onChange={(environmentBundleRefs) => setWorkspace({
          environmentBundleRefs,
        })}
      />
      <WorkerTemplateConfigDocumentBindingsField
        requirements={configDocumentRequirements}
        value={draft.spec.workspace.configDocumentBindings}
        catalog={catalog}
        onChange={(configDocumentBindings) => setWorkspace({
          configDocumentBindings,
        })}
      />
      <FormField
        label={t("fields.instructions")}
        htmlFor="worker-instructions"
      >
        <Textarea
          id="worker-instructions"
          rows={6}
          value={draft.spec.workspace.instructions}
          onChange={(event) => setWorkspace({
            instructions: event.target.value,
          })}
        />
      </FormField>
    </FormFieldGroup>
  );
}
