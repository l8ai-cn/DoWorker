"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";
import { buildDocsRows, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

export default function ChannelsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.features.channels.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.features.channels.description")}
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.overview.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.channels.overview.description")}
        </p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.channels.overview.item1")}</li>
          <li>{t("docs.features.channels.overview.item2")}</li>
          <li>{t("docs.features.channels.overview.item3")}</li>
          <li>{t("docs.features.channels.overview.item4")}</li>
        </ul>
      </section>

      {/* Creating Channels */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.creating.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.channels.creating.description")}
        </p>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.features.channels.creating", "fieldHeader", "descriptionHeader")}
          rows={buildDocsRows(t, "docs.features.channels.creating", [
            "name",
            "channelDescription",
            "projectId",
            "ticketSlug",
          ])}
        />
      </section>

      {/* Message Types */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.messageTypes.title")}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              <code className="bg-muted px-1 rounded">
                {t("docs.features.channels.messageTypes.text")}
              </code>
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.messageTypes.textDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              <code className="bg-muted px-1 rounded">
                {t("docs.features.channels.messageTypes.system")}
              </code>
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.messageTypes.systemDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Mentions */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.mentions.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.channels.mentions.description")}
        </p>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm">
          <pre className="text-success">{`// Send a message with mentions
send_channel_message({
  channel_id: 123,
  content: "Can you review this implementation?",
  message_type: "text",
  mentions: ["pod-abc", "pod-xyz"]
})`}</pre>
        </div>
        <p className="text-sm text-muted-foreground mt-4">
          {t("docs.features.channels.mentions.hint", {
            param: "mentioned_pod",
          })}
        </p>
      </section>

      {/* Shared Documents */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.sharedDocs.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.channels.sharedDocs.description")}
        </p>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.channels.sharedDocs.getDoc")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.sharedDocs.getDocDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.channels.sharedDocs.updateDoc")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.sharedDocs.updateDocDesc")}
            </p>
          </div>
        </div>
        <p className="text-sm text-muted-foreground mt-4">
          {t("docs.features.channels.sharedDocs.hint")}
        </p>
      </section>

      {/* MCP Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.mcpTools.title")}
        </h2>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.features.channels.mcpTools", "toolHeader", "descriptionHeader")}
          rows={buildDocsRows(t, "docs.features.channels.mcpTools", [
            "searchChannels",
            "createChannel",
            "getChannel",
            "sendMessage",
            "getMessages",
            "getDocument",
            "updateDocument",
          ])}
        />
      </section>

      {/* Use Cases */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.useCases.title")}
        </h2>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.channels.useCases.coordination")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.useCases.coordinationDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.channels.useCases.designDocs")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.useCases.designDocsDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.features.channels.useCases.notifications")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.features.channels.useCases.notificationsDesc")}
            </p>
          </div>
        </div>
      </section>

      {/* Web UI */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.features.channels.webUI.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.features.channels.webUI.description")}
        </p>
        <ol className="list-decimal list-inside text-muted-foreground space-y-2">
          <li>{t("docs.features.channels.webUI.step1")}</li>
          <li>{t("docs.features.channels.webUI.step2")}</li>
          <li>{t("docs.features.channels.webUI.step3")}</li>
          <li>{t("docs.features.channels.webUI.step4")}</li>
          <li>{t("docs.features.channels.webUI.step5")}</li>
        </ol>
      </section>

      <DocNavigation />
    </div>
  );
}
