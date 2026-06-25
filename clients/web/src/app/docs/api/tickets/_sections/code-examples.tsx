"use client";

import { useTranslations } from "next-intl";

export function CodeExamples() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">
        {t("docs.api.tickets.examples.title")}
      </h2>
      <div className="space-y-4">
        <div className="surface-card p-4">
          <h3 className="font-medium mb-3">
            {t("docs.api.tickets.examples.listTickets")}
          </h3>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm">
            <pre className="text-success">{`curl -X GET \\
  "https://your-domain.com/api/v1/ext/orgs/my-org/tickets" \\
  -H "X-API-Key: amk_your_api_key_here"`}</pre>
          </div>
        </div>
        <div className="surface-card p-4">
          <h3 className="font-medium mb-3">
            {t("docs.api.tickets.examples.createTicket")}
          </h3>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm">
            <pre className="text-success">{`curl -X POST \\
  "https://your-domain.com/api/v1/ext/orgs/my-org/tickets" \\
  -H "X-API-Key: amk_your_api_key_here" \\
  -H "Content-Type: application/json" \\
  -d '{
    "title": "Implement user auth",
    "type": "feature",
    "priority": "high"
  }'`}</pre>
          </div>
        </div>
      </div>
    </section>
  );
}
