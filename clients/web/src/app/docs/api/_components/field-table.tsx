"use client";

import { useTranslations } from "next-intl";
import { DocsTable } from "@/components/docs/DocsTable";
import { docsMono, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

interface FieldTableProps {
  rows: Array<{ field: string; descKey: string }>;
}

export function FieldTable({ rows }: FieldTableProps) {
  const t = useTranslations();
  const prefix = "docs.api.common";

  return (
    <DocsTable
      columns={twoColumnHeaders(t, prefix, "fieldHeader", "descriptionHeader")}
      rows={rows.map((row) => ({
        cells: [docsMono(row.field), t(row.descKey)],
      }))}
    />
  );
}
