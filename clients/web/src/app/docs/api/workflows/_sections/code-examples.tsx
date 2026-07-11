"use client";

import { useTranslations } from "next-intl";

export function CodeExamples() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">
        {t("docs.api.workflows.examples.title")}
      </h2>

      <div className="space-y-6">
        <div>
          <h3 className="font-medium mb-2">
            {t("docs.api.workflows.examples.createWorkflow")}
          </h3>
          <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm overflow-x-auto">
            <code>{`curl -X POST /api/v1/orgs/{org}/workflows \\
  -H "Authorization: Bearer {token}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Daily Code Review",
    "agent_slug": 1,
    "prompt_template": "Review changes in {{branch}} branch",
    "prompt_variables": {"branch": "main"},
    "execution_mode": "autopilot",
    "cron_expression": "0 9 * * *",
    "sandbox_strategy": "persistent",
    "timeout_minutes": 30
  }'`}</code>
          </pre>
        </div>

        <div>
          <h3 className="font-medium mb-2">
            {t("docs.api.workflows.examples.triggerRun")}
          </h3>
          <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm overflow-x-auto">
            <code>{`curl -X POST /api/v1/orgs/{org}/workflows/{slug}/trigger \\
  -H "Authorization: Bearer {token}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "variables": {
      "branch": "feature/new-api",
      "focus_area": "security"
    }
  }'`}</code>
          </pre>
        </div>
      </div>
    </section>
  );
}
