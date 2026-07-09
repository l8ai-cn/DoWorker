"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";
import { buildDocsRows, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

export default function AgentPodPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.features.agentpod.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.features.agentpod.description")}
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.agentpod.overview.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.agentpod.overview.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
          <li>
            <strong>{t("docs.features.agentpod.overview.terminal").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.agentpod.overview.terminal").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.agentpod.overview.worktree").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.agentpod.overview.worktree").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.agentpod.overview.agent").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.agentpod.overview.agent").split(" — ")[1]}
          </li>
          <li>
            <strong>{t("docs.features.agentpod.overview.mcpTools").split(" — ")[0]}</strong>
            {" — "}
            {t("docs.features.agentpod.overview.mcpTools").split(" — ")[1]}
          </li>
        </ul>
      </section>

      {/* Features */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.agentpod.keyFeatures.title")}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.webTerminal")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.webTerminalDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.statusMonitoring")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.statusMonitoringDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.worktreeIsolation")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.worktreeIsolationDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.multiplePods")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.multiplePodsDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.ticketIntegration")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.ticketIntegrationDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.agentpod.keyFeatures.autoCleanup")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.agentpod.keyFeatures.autoCleanupDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Pod Lifecycle */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.agentpod.lifecycle.title")}
        </h2>
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-primary">
              {t("docs.features.agentpod.lifecycle.initializing")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.initializingDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-success">
              {t("docs.features.agentpod.lifecycle.running")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.runningDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-warning">
              {t("docs.features.agentpod.lifecycle.paused")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.pausedDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-warning">
              {t("docs.features.agentpod.lifecycle.disconnected")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.disconnectedDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.completed")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.completedDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-danger">
              {t("docs.features.agentpod.lifecycle.terminated")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.terminatedDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-primary">
              {t("docs.features.agentpod.lifecycle.orphaned")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.orphanedDesc")}
            </p>
          </div>
          <div className="flex items-start gap-4">
            <div className="w-24 text-sm font-medium text-danger">
              {t("docs.features.agentpod.lifecycle.error")}
            </div>
            <p className="text-muted-foreground">
              {t("docs.features.agentpod.lifecycle.errorDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Agent Status */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.agentpod.agentStatus.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.agentpod.agentStatus.description")}
        </p>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.features.agentpod.agentStatus", "statusHeader", "descriptionHeader")}
          rows={buildDocsRows(t, "docs.features.agentpod.agentStatus", ["idle", "executing", "waiting"])}
        />
      </section>

      {/* Configuration */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.agentpod.configuration.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.features.agentpod.configuration.description")}
        </p>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.features.agentpod.configuration", "optionHeader", "descriptionHeader")}
          rows={buildDocsRows(t, "docs.features.agentpod.configuration", [
            "agent",
            "model",
            "automationLevel",
            "repository",
            "ticket",
            "prompt",
          ])}
        />
      </section>

      <DocNavigation />
    </div>
  );
}
