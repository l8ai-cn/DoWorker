"use client";

import { useTranslations } from "next-intl";
import { createLoopFields } from "./endpoints-data";
import { FieldTable } from "../../_components/field-table";

export function EndpointDetails() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-6">{t("docs.api.loops.details.title")}</h2>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.listLoops.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.loops.details.listLoops.description")}
        </p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.createLoop.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.loops.details.createLoop.description")}
        </p>
        <h4 className="font-medium mb-2">{t("docs.api.common.requestBody")}</h4>
        <FieldTable
          rows={createLoopFields.map((field) => ({
            field,
            descKey: `docs.api.loops.details.createLoop.fields.${field}`,
          }))}
        />
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.getLoop.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.getLoop.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.updateLoop.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.updateLoop.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.deleteLoop.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.deleteLoop.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.enableLoop.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.loops.details.enableLoop.description")}
        </p>
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.disableLoop.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.disableLoop.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.triggerLoop.title")}
        </h3>
        <p className="text-muted-foreground mb-4">
          {t("docs.api.loops.details.triggerLoop.description")}
        </p>
        <h4 className="font-medium mb-2">{t("docs.api.common.requestBody")}</h4>
        <FieldTable
          rows={[
            {
              field: "variables",
              descKey: "docs.api.loops.details.triggerLoop.fields.variables",
            },
          ]}
        />
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.listRuns.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.listRuns.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.getRun.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.getRun.description")}</p>
      </div>

      <div className="mb-8 surface-card p-6">
        <h3 className="text-lg font-semibold mb-2 font-mono">
          {t("docs.api.loops.details.cancelRun.title")}
        </h3>
        <p className="text-muted-foreground">{t("docs.api.loops.details.cancelRun.description")}</p>
      </div>
    </section>
  );
}
