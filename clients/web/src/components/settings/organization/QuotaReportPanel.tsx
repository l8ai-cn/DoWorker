"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { type QuotaReport, type ScopeUsage, getQuotaReport } from "@/lib/api/quotaApi";

function usd(v: number): string {
  return `$${v.toFixed(2)}`;
}

function Row({ label, u }: { label: string; u: ScopeUsage }) {
  return (
    <div className="flex items-center justify-between py-1.5 text-sm">
      <span>{label}</span>
      <span className={u.over_limit ? "text-destructive" : "text-muted-foreground"}>
        {u.tokens.toLocaleString()} tokens · {usd(u.cost_usd)}
        {u.limit_tokens != null && ` / ${u.limit_tokens.toLocaleString()}`}
        {u.over_limit && " · OVER"}
      </span>
    </div>
  );
}

export function QuotaReportPanel() {
  const [report, setReport] = useState<QuotaReport | null>(null);

  useEffect(() => {
    getQuotaReport()
      .then(setReport)
      .catch((e) => toast.error(e instanceof Error ? e.message : "Failed to load report"));
  }, []);

  if (!report) {
    return (
      <div className="surface-card p-6 text-sm text-muted-foreground">Loading usage…</div>
    );
  }

  return (
    <div className="surface-card space-y-4 p-6">
      <div className="flex items-baseline justify-between">
        <h2 className="text-lg font-semibold">Usage &amp; Billing</h2>
        <span className="text-sm text-muted-foreground">
          {report.total_tokens.toLocaleString()} tokens · {usd(report.total_cost_usd)}
        </span>
      </div>

      {report.quotas.length > 0 && (
        <section>
          <h3 className="mb-1 text-sm font-medium">Quota status</h3>
          <div className="divide-y divide-border">
            {report.quotas.map((q, i) => (
              <Row
                key={`q${i}`}
                label={`${q.user_id ? `User ${q.user_id}` : "Org"}${q.model ? ` · ${q.model}` : ""}`}
                u={q}
              />
            ))}
          </div>
        </section>
      )}

      <section>
        <h3 className="mb-1 text-sm font-medium">By model</h3>
        <div className="divide-y divide-border">
          {report.by_model.map((m, i) => (
            <Row key={`m${i}`} label={m.model ?? "unknown"} u={m} />
          ))}
          {report.by_model.length === 0 && (
            <p className="py-1.5 text-sm text-muted-foreground">No usage recorded.</p>
          )}
        </div>
      </section>

      <section>
        <h3 className="mb-1 text-sm font-medium">By user</h3>
        <div className="divide-y divide-border">
          {report.by_user.map((u, i) => (
            <Row key={`u${i}`} label={`User ${u.user_id ?? "?"}`} u={u} />
          ))}
          {report.by_user.length === 0 && (
            <p className="py-1.5 text-sm text-muted-foreground">No usage recorded.</p>
          )}
        </div>
      </section>
    </div>
  );
}
