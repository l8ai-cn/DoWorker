"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";

export default function LoopsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.features.loops.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.features.loops.description")}
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.overview.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.loops.overview.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.loops.overview.item1")}</li>
          <li>{t("docs.features.loops.overview.item2")}</li>
          <li>{t("docs.features.loops.overview.item3")}</li>
          <li>{t("docs.features.loops.overview.item4")}</li>
          <li>{t("docs.features.loops.overview.item5")}</li>
          <li>{t("docs.features.loops.overview.item6")}</li>
        </ul>
      </section>

      {/* Execution Modes */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.executionModes.title")}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.executionModes.autopilot")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.executionModes.autopilotDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.executionModes.direct")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.executionModes.directDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Trigger Types */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.triggerTypes.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.loops.triggerTypes.description")}
        </p>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.triggerTypes.cron")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.triggerTypes.cronDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.triggerTypes.api")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.triggerTypes.apiDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Sandbox Strategies */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.sandboxStrategies.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.loops.sandboxStrategies.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.sandboxStrategies.persistent")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.sandboxStrategies.persistentDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.loops.sandboxStrategies.fresh")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.loops.sandboxStrategies.freshDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Session Persistence */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.sessionPersistence.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed">
          {t("docs.features.loops.sessionPersistence.description")}
        </p>
      </section>

      {/* Concurrency Policies */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.concurrency.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.loops.concurrency.description")}
        </p>
        <DocsTable
          columns={[
            { header: t("docs.features.loops.concurrency.modeHeader"), className: "w-1/4" },
            { header: t("docs.features.loops.concurrency.behaviorHeader") },
          ]}
          rows={[
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.loops.concurrency.skip")}</span>,
                t("docs.features.loops.concurrency.skipDesc"),
              ],
            },
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.loops.concurrency.queue")}</span>,
                t("docs.features.loops.concurrency.queueDesc"),
              ],
            },
            {
              cells: [
                <span key="m" className="font-medium">{t("docs.features.loops.concurrency.replace")}</span>,
                t("docs.features.loops.concurrency.replaceDesc"),
              ],
            },
          ]}
        />
      </section>

      {/* Prompt Templates */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.promptTemplates.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.loops.promptTemplates.description")}
        </p>
        <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm overflow-x-auto">
          <code>{t("docs.features.loops.promptTemplates.example")}</code>
        </pre>
      </section>

      {/* Webhook Callbacks */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.webhookCallbacks.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed">
          {t("docs.features.loops.webhookCallbacks.description")}
        </p>
      </section>

      {/* Use Cases */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.loops.useCases.title")}
        </h2>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.loops.useCases.item1")}</li>
          <li>{t("docs.features.loops.useCases.item2")}</li>
          <li>{t("docs.features.loops.useCases.item3")}</li>
          <li>{t("docs.features.loops.useCases.item4")}</li>
          <li>{t("docs.features.loops.useCases.item5")}</li>
        </ul>
      </section>

      <DocNavigation />
    </div>
  );
}
