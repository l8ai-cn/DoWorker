"use client";

import { useTranslations } from "next-intl";
import { MethodBadge } from "./method-badge";

export type EndpointRow = {
  method: string;
  path: string;
  scope: string;
  descKey: string;
};

export function EndpointSummaryTable({ rows }: { rows: EndpointRow[] }) {
  const t = useTranslations();

  return (
    <div className="overflow-x-auto rounded-lg surface-card">
      <table className="w-full text-sm divide-y divide-border/20">
        <thead>
          <tr className="bg-surface-muted/50">
            <th className="text-left p-3">
              {t("docs.api.common.methodHeader")}
            </th>
            <th className="text-left p-3">
              {t("docs.api.common.pathHeader")}
            </th>
            <th className="text-left p-3">
              {t("docs.api.common.scopeHeader")}
            </th>
            <th className="text-left p-3">
              {t("docs.api.common.descriptionHeader")}
            </th>
          </tr>
        </thead>
        <tbody className="text-muted-foreground divide-y divide-border/20">
          {rows.map((row) => (
            <tr key={`${row.method}-${row.path}`}>
              <td className="p-3">
                <MethodBadge method={row.method} />
              </td>
              <td className="p-3 font-mono text-xs">{row.path}</td>
              <td className="p-3 font-mono text-xs">{row.scope}</td>
              <td className="p-3">{t(row.descKey)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
