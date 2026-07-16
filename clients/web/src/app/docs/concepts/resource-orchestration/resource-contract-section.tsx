"use client";

import { useTranslations } from "next-intl";
import { DocsTable } from "@/components/docs/DocsTable";
import { lifecycleItems, resourceKindRows } from "./resource-orchestration-data";

export function ResourceContractSection() {
  const t = useTranslations("resourceOrchestration");
  const rows = resourceKindRows.map((row) => ({
    cells: [t(row.nameKey), t(row.applyKey), t(row.purposeKey)],
  }));

  return (
    <>
      <section className="mb-12">
        <h2 className="text-2xl font-semibold text-foreground">
          {t("lifecycleTitle")}
        </h2>
        <p className="mt-3 max-w-3xl leading-relaxed text-muted-foreground">
          {t("lifecycleDescription")}
        </p>
        <ol className="mt-6 grid gap-5 border-y border-border/60 py-6 md:grid-cols-3">
          {lifecycleItems.map((item, index) => (
            <li key={item.titleKey} className="min-w-0">
              <p className="text-xs font-semibold text-primary">
                {String(index + 1).padStart(2, "0")}
              </p>
              <h3 className="mt-2 font-semibold text-foreground">
                {t(item.titleKey)}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {t(item.descriptionKey)}
              </p>
            </li>
          ))}
        </ol>
      </section>

      <section className="mb-12">
        <h2 className="mb-4 text-2xl font-semibold text-foreground">
          {t("kindsTitle")}
        </h2>
        <DocsTable
          columns={[
            { header: t("kindsColumns.kind") },
            { header: t("kindsColumns.apply") },
            { header: t("kindsColumns.purpose") },
          ]}
          rows={rows}
        />
      </section>

      <section className="mb-12">
        <h2 className="text-2xl font-semibold text-foreground">
          {t("referencesTitle")}
        </h2>
        <p className="mt-3 max-w-3xl leading-relaxed text-muted-foreground">
          {t("referencesDescription")}
        </p>
        <pre className="mt-5 overflow-x-auto rounded-lg bg-surface-muted p-4 text-sm text-foreground">
          <code>{`workerTemplateRef:
  kind: WorkerTemplate
  name: codex-reviewer
  revision: 3`}</code>
        </pre>
        <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
          {t("referencesPinned")}
        </p>
      </section>
    </>
  );
}
