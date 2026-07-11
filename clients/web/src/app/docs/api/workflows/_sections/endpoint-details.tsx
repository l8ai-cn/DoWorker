"use client";

import { useTranslations } from "next-intl";
import { createWorkflowFields } from "./endpoints-data";
import { FieldTable } from "../../_components/field-table";

export function EndpointDetails() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-6">{t("docs.api.workflows.details.title")}</h2>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.listWorkflows.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.workflows.details.listWorkflows.description")}
        </p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.createWorkflow.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.workflows.details.createWorkflow.description")}
        </p>
        <h4 className="font-medium mb-2">{t("docs.api.common.requestBody")}</h4>
        <FieldTable
          rows={createWorkflowFields.map((field) => ({
            field,
            descKey: `docs.api.workflows.details.createWorkflow.fields.${field}`,
          }))}
        />
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.getWorkflow.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.getWorkflow.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.updateWorkflow.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.updateWorkflow.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.deleteWorkflow.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.deleteWorkflow.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.enableWorkflow.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.workflows.details.enableWorkflow.description")}
        </p>
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.disableWorkflow.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.disableWorkflow.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.triggerWorkflow.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.workflows.details.triggerWorkflow.description")}
        </p>
        <h4 className="font-medium mb-2">{t("docs.api.common.requestBody")}</h4>
        <FieldTable
          rows={[
            {
              field: "variables",
              descKey: "docs.api.workflows.details.triggerWorkflow.fields.variables",
            },
          ]}
        />
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.listRuns.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.listRuns.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.getRun.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.getRun.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.workflows.details.cancelRun.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.workflows.details.cancelRun.description")}</p>
      </div>
    </section>
  );
}
