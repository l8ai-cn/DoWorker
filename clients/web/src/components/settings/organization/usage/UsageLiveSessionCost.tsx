"use client";

import type { OrgLiveUsageSummary } from "@/lib/api/orgLiveUsageFetch";
import type { TranslationFn } from "../GeneralSettings";
import { formatUsd } from "./format";

interface UsageLiveSessionCostProps {
  live: OrgLiveUsageSummary | null;
  t: TranslationFn;
}

export function UsageLiveSessionCost({ live, t }: UsageLiveSessionCostProps) {
  const models = live?.usage_by_model ? Object.values(live.usage_by_model) : [];
  const sorted = [...models].sort(
    (a, b) => (b.total_cost_usd ?? 0) - (a.total_cost_usd ?? 0),
  );

  return (
    <div className="surface-card p-6">
      <div className="flex flex-wrap items-baseline justify-between gap-2 mb-4">
        <div>
          <h3 className="text-sm font-medium">{t("settings.usagePage.liveSessionsTitle")}</h3>
          <p className="text-xs text-muted-foreground mt-1">
            {t("settings.usagePage.liveSessionsHint")}
          </p>
        </div>
        {typeof live?.total_cost_usd === "number" && (
          <p className="text-2xl font-bold text-foreground" title={formatUsd(live.total_cost_usd, true)}>
            {formatUsd(live.total_cost_usd)}
          </p>
        )}
      </div>
      {sorted.length === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-4">
          {t("settings.usagePage.liveSessionsEmpty")}
        </p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <caption className="sr-only">{t("settings.usagePage.liveSessionsTitle")}</caption>
            <thead>
              <tr className="border-b border-border text-left">
                <th scope="col" className="pb-2 font-medium text-muted-foreground">
                  {t("settings.usagePage.columnModel")}
                </th>
                <th scope="col" className="pb-2 font-medium text-muted-foreground text-right">
                  {t("settings.usagePage.estimatedCost")}
                </th>
              </tr>
            </thead>
            <tbody>
              {sorted.map((row) => (
                <tr key={row.model} className="border-b border-border/50 last:border-0">
                  <td className="py-2 font-medium">{row.model}</td>
                  <td className="py-2 text-right font-medium">
                    {typeof row.total_cost_usd === "number"
                      ? formatUsd(row.total_cost_usd)
                      : "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
