"use client";

import { useTranslations } from "next-intl";
import { EndpointSummaryTable } from "../../_components/endpoint-summary-table";
import { summaryRows } from "./endpoints-data";

export function EndpointsTable() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">
        {t("docs.api.workflows.endpoints.title")}
      </h2>
      <EndpointSummaryTable rows={summaryRows} />
    </section>
  );
}
