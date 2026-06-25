"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";
import { docsLabel, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

export default function TicketsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.features.tickets.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.features.tickets.description")}
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.overview.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.tickets.overview.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.tickets.overview.item1")}</li>
          <li>{t("docs.features.tickets.overview.item2")}</li>
          <li>{t("docs.features.tickets.overview.item3")}</li>
          <li>{t("docs.features.tickets.overview.item4")}</li>
          <li>{t("docs.features.tickets.overview.item5")}</li>
        </ul>
      </section>

      {/* Ticket Status */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.status.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.tickets.status.description")}
        </p>
        <div className="flex flex-wrap gap-2">
          <span className="px-3 py-1 bg-muted rounded text-sm">
            {t("docs.features.tickets.status.backlog")}
          </span>
          <span className="text-muted-foreground">&rarr;</span>
          <span className="px-3 py-1 bg-muted rounded text-sm">
            {t("docs.features.tickets.status.todo")}
          </span>
          <span className="text-muted-foreground">&rarr;</span>
          <span className="px-3 py-1 bg-info/20 text-info rounded text-sm">
            {t("docs.features.tickets.status.inProgress")}
          </span>
          <span className="text-muted-foreground">&rarr;</span>
          <span className="px-3 py-1 bg-warning/20 text-warning rounded text-sm">
            {t("docs.features.tickets.status.inReview")}
          </span>
          <span className="text-muted-foreground">&rarr;</span>
          <span className="px-3 py-1 bg-success/20 text-success rounded text-sm">
            {t("docs.features.tickets.status.done")}
          </span>
        </div>
      </section>

      {/* Priority */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.priority.title")}
        </h2>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.features.tickets.priority", "priorityHeader", "descriptionHeader")}
          rows={[
            { cells: [<span key="u" className="font-medium text-danger">{t("docs.features.tickets.priority.urgent")}</span>, t("docs.features.tickets.priority.urgentDesc")] },
            { cells: [<span key="h" className="font-medium text-warning">{t("docs.features.tickets.priority.high")}</span>, t("docs.features.tickets.priority.highDesc")] },
            { cells: [<span key="m" className="font-medium text-warning">{t("docs.features.tickets.priority.medium")}</span>, t("docs.features.tickets.priority.mediumDesc")] },
            { cells: [<span key="l" className="font-medium text-muted-foreground">{t("docs.features.tickets.priority.low")}</span>, t("docs.features.tickets.priority.lowDesc")] },
            { cells: [docsLabel(t("docs.features.tickets.priority.none")), t("docs.features.tickets.priority.noneDesc")] },
          ]}
        />
      </section>

      {/* Pod Integration */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.podIntegration.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.tickets.podIntegration.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>
            <strong>{t("docs.features.tickets.podIntegration.context").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.tickets.podIntegration.context").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.tickets.podIntegration.progress").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.tickets.podIntegration.progress").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.tickets.podIntegration.autoUpdate").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.tickets.podIntegration.autoUpdate").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.tickets.podIntegration.history").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.tickets.podIntegration.history").split(" — ")[1]}
          </li>
        </ul>
      </section>

      {/* Git Integration */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.gitIntegration.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.tickets.gitIntegration.description")}
        </p>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.tickets.gitIntegration.commits")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.tickets.gitIntegration.commitsDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.tickets.gitIntegration.mrs")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.tickets.gitIntegration.mrsDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Estimation */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.tickets.storyPoints.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.tickets.storyPoints.description")}
        </p>
        <div className="flex flex-wrap gap-2">
          {[1, 2, 3, 5, 8, 13, 21].map((point) => (
            <span
              key={point}
              className="w-10 h-10 flex items-center justify-center bg-muted rounded text-sm font-medium"
            >
              {point}
            </span>
          ))}
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}
