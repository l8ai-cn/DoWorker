"use client";

import { useTranslations } from "next-intl";
import { DocsHorizontalScroll } from "@/components/docs/DocsHorizontalScroll";
import { RequiredBadge, OptionalBadge } from "./method-badge";

export type ParamRow = {
  name: string;
  type: string;
  required: boolean;
  default?: string;
  descKey?: string;
  desc?: string;
};

export type TableKind = "query" | "path" | "body";

const KIND_TITLE: Record<TableKind, string> = {
  query: "docs.api.common.queryParams",
  path: "docs.api.common.pathParams",
  body: "docs.api.common.requestBody",
};

const KIND_FIRST_HEADER: Record<TableKind, string> = {
  query: "docs.api.common.paramHeader",
  path: "docs.api.common.paramHeader",
  body: "docs.api.common.fieldHeader",
};

export function ParamTable({
  kind,
  rows,
  withDefault = false,
}: {
  kind: TableKind;
  rows: ParamRow[];
  withDefault?: boolean;
}) {
  const t = useTranslations();

  return (
    <div>
      <h4 className="font-medium mb-2 text-foreground">{t(KIND_TITLE[kind])}</h4>
      <DocsHorizontalScroll>
        <div className="rounded-lg surface-card overflow-hidden">
          <table
            className={`${withDefault ? "min-w-[820px]" : "min-w-[720px]"} text-sm divide-y divide-border/20`}
          >
            <thead>
              <tr className="bg-surface-muted/50">
                <th className="text-left p-3">
                  {t(KIND_FIRST_HEADER[kind])}
                </th>
                <th className="text-left p-3">
                  {t("docs.api.common.typeHeader")}
                </th>
                <th className="text-left p-3">
                  {t("docs.api.common.requiredHeader")}
                </th>
                {withDefault && (
                  <th className="text-left p-3">
                    {t("docs.api.common.defaultHeader")}
                  </th>
                )}
                <th className="text-left p-3">
                  {t("docs.api.common.descriptionHeader")}
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground divide-y divide-border/20">
              {rows.map((row) => (
                <tr key={row.name}>
                  <td className="p-3 font-mono text-xs">{row.name}</td>
                  <td className="p-3">{row.type}</td>
                  <td className="p-3">
                    {row.required ? <RequiredBadge /> : <OptionalBadge />}
                  </td>
                  {withDefault && (
                    <td className="p-3">{row.default ?? "-"}</td>
                  )}
                  <td className="p-3">
                    {row.desc !== undefined ? row.desc : t(row.descKey ?? "")}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </DocsHorizontalScroll>
    </div>
  );
}
