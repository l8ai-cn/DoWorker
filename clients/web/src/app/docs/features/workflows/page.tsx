"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";

export default function WorkflowsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.features.workflows.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.features.workflows.description")}
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.overview.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.workflows.overview.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.workflows.overview.item1")}</li>
          <li>{t("docs.features.workflows.overview.item2")}</li>
          <li>{t("docs.features.workflows.overview.item3")}</li>
          <li>{t("docs.features.workflows.overview.item4")}</li>
          <li>{t("docs.features.workflows.overview.item5")}</li>
          <li>{t("docs.features.workflows.overview.item6")}</li>
        </ul>
      </section>

      {/* Execution Modes */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.executionModes.title")}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.executionModes.autopilot")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.executionModes.autopilotDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.executionModes.direct")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.executionModes.directDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Trigger Types */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.triggerTypes.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.workflows.triggerTypes.description")}
        </p>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.triggerTypes.cron")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.triggerTypes.cronDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.triggerTypes.api")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.triggerTypes.apiDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Sandbox Strategies */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.sandboxStrategies.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.workflows.sandboxStrategies.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.sandboxStrategies.persistent")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.sandboxStrategies.persistentDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.workflows.sandboxStrategies.fresh")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.workflows.sandboxStrategies.freshDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Session Persistence */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.sessionPersistence.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed">
          {t("docs.features.workflows.sessionPersistence.description")}
        </p>
      </section>

      {/* Concurrency Policies */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.concurrency.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.workflows.concurrency.description")}
        </p>
        <DocsTable
          columns={[
            { header: t("docs.features.workflows.concurrency.modeHeader"), className: "w-1/4" },
            { header: t("docs.features.workflows.concurrency.behaviorHeader") },
          ]}
          rows={[
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.workflows.concurrency.skip")}</span>,
                t("docs.features.workflows.concurrency.skipDesc"),
              ],
            },
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.workflows.concurrency.queue")}</span>,
                t("docs.features.workflows.concurrency.queueDesc"),
              ],
            },
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.workflows.concurrency.replace")}</span>,
                t("docs.features.workflows.concurrency.replaceDesc"),
              ],
            },
          ]}
        />
      </section>

      {/* Prompt Templates */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.promptTemplates.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.workflows.promptTemplates.description")}
        </p>
        <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm overflow-x-auto">
          <code>{t("docs.features.workflows.promptTemplates.example")}</code>
        </pre>
      </section>

      {/* Webhook Callbacks */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.webhookCallbacks.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed">
          {t("docs.features.workflows.webhookCallbacks.description")}
        </p>
      </section>

      {/* Use Cases */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.workflows.useCases.title")}
        </h2>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.workflows.useCases.item1")}</li>
          <li>{t("docs.features.workflows.useCases.item2")}</li>
          <li>{t("docs.features.workflows.useCases.item3")}</li>
          <li>{t("docs.features.workflows.useCases.item4")}</li>
          <li>{t("docs.features.workflows.useCases.item5")}</li>
        </ul>
      </section>

      <DocNavigation />
    </div>
  );
}
