"use client";

import { useEffect, useState } from "react";
import { getQuotaReport, type ScopeUsage } from "@/lib/api/quotaApi";

interface Props {
  tokenBudget: number | null;
  onChange: (budget: number | null) => void;
  t: (key: string) => string;
}

function formatTokens(n: number): string {
  return n.toLocaleString();
}

/**
 * Optional per-Worker budget mount. The numeric cap is emitted as
 * `CONFIG token_budget` in the AgentFile layer (existing backend contract);
 * the org-level usage/quota is shown read-only so the operator can see the
 * quota gate that the backend enforces at pod creation.
 */
export function WorkerBudgetSection({ tokenBudget, onChange, t }: Props) {
  const [orgUsage, setOrgUsage] = useState<ScopeUsage | null>(null);

  useEffect(() => {
    let alive = true;
    getQuotaReport()
      .then((report) => {
        if (!alive) return;
        const orgQuota = report.quotas.find((q) => q.user_id == null);
        setOrgUsage(
          orgQuota ?? {
            tokens: report.total_tokens,
            cost_usd: report.total_cost_usd,
            over_limit: false,
          },
        );
      })
      .catch(() => {
        // Quota endpoint is best-effort here; the budget input stays usable.
      });
    return () => {
      alive = false;
    };
  }, []);

  const handleInput = (raw: string) => {
    const trimmed = raw.trim();
    if (trimmed === "") {
      onChange(null);
      return;
    }
    const parsed = Number.parseInt(trimmed, 10);
    onChange(Number.isFinite(parsed) && parsed > 0 ? parsed : null);
  };

  return (
    <div className="space-y-3">
      {orgUsage && (
        <div className="rounded-md border border-border bg-muted/30 p-3 text-xs">
          <div className="flex items-center justify-between">
            <span className="font-medium">{t("ide.createPod.budgetOrgUsageLabel")}</span>
            <span
              className={
                orgUsage.over_limit ? "text-destructive" : "text-muted-foreground"
              }
            >
              {formatTokens(orgUsage.tokens)}
              {orgUsage.limit_tokens != null &&
                ` / ${formatTokens(orgUsage.limit_tokens)}`}
              {" "}
              {t("ide.createPod.budgetTokensUnit")}
            </span>
          </div>
          <p className="mt-1 text-muted-foreground">
            {t("ide.createPod.budgetOrgUsageHint")}
          </p>
        </div>
      )}

      <label className="block">
        <span className="mb-1 block text-sm font-medium">
          {t("ide.createPod.budgetCapLabel")}
        </span>
        <input
          type="number"
          min={1}
          step={1000}
          inputMode="numeric"
          data-testid="worker-token-budget-input"
          value={tokenBudget ?? ""}
          onChange={(e) => handleInput(e.target.value)}
          placeholder={t("ide.createPod.budgetCapPlaceholder")}
          className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/30"
        />
      </label>
      <p className="text-xs text-muted-foreground">
        {t("ide.createPod.budgetCapHint")}
      </p>
    </div>
  );
}

export default WorkerBudgetSection;
