"use client";

import { useTranslations } from "next-intl";
import { MethodBadge } from "./method-badge";
import { JsonBlock } from "./code-block";
import { ParamTable, type ParamRow, type TableKind } from "./param-table";

export type TableSpec = {
  kind: TableKind;
  withDefault?: boolean;
  rows: ParamRow[];
};

export type DetailSpec = {
  method: string;
  path: string;
  descKey: string;
  tables: TableSpec[];
  response: string;
};

export function EndpointDetail({
  spec,
  pathAsCode,
  jsonTone,
}: {
  spec: DetailSpec;
  pathAsCode?: boolean;
  jsonTone?: "success";
}) {
  const t = useTranslations();

  return (
    <div className="surface-card p-6 space-y-6">
      {pathAsCode ? (
        <div>
          <h3 className="text-lg font-semibold mb-2">
            <MethodBadge method={spec.method} size="md" />{" "}
            <code className="text-sm">{spec.path}</code>
          </h3>
          <p className="text-muted-foreground text-sm">{t(spec.descKey)}</p>
        </div>
      ) : (
        <>
          <h3 className="text-lg font-semibold">
            <MethodBadge method={spec.method} size="md" />{" "}
            {spec.path}
          </h3>
          <p className="text-muted-foreground">{t(spec.descKey)}</p>
        </>
      )}

      {spec.tables.map((table, index) => (
        <ParamTable key={index} {...table} />
      ))}

      <div>
        <h4 className="font-medium mb-2">
          {t("docs.api.common.responseExample")}
        </h4>
        <JsonBlock tone={jsonTone}>{spec.response}</JsonBlock>
      </div>
    </div>
  );
}
