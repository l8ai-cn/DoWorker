"use client";

import { useTranslations } from "next-intl";
import { buildDocsRows, threeColumnHeaders, twoColumnHeaders } from "@/components/docs/docs-table-helpers";
import { McpToolTableSection } from "./McpToolTableSection";

const PREFIX = "docs.runners.mcpTools";

export function OverviewSection() {
  const t = useTranslations();
  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">
        {t(`${PREFIX}.overview.title`)}
      </h2>
      <p className="text-muted-foreground leading-relaxed mb-4">
        {t(`${PREFIX}.overview.description`)}
      </p>
      <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
        <p className="text-sm text-muted-foreground">
          <strong>{t(`${PREFIX}.overview.autoConfig`)}</strong>
        </p>
      </div>
    </section>
  );
}

export function DiscoveryToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.discovery`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      description={t(`${prefix}.description`)}
      columns={twoColumnHeaders(t, prefix, "toolHeader", "descriptionHeader")}
      rows={buildDocsRows(t, prefix, ["listPods", "listRunners", "listRepos"])}
    />
  );
}

export function PodToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.pod`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      description={t(`${prefix}.description`)}
      columns={threeColumnHeaders(t, prefix, "toolHeader", "descriptionHeader", "paramsHeader")}
      rows={buildDocsRows(t, prefix, ["createPod", "getPodSnapshot", "sendPodInput", "getPodStatus"], { params: true })}
    />
  );
}

export function BindingToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.binding`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      columns={threeColumnHeaders(t, prefix, "toolHeader", "descriptionHeader", "paramsHeader")}
      rows={buildDocsRows(
        t,
        prefix,
        ["bindPod", "acceptBinding", "rejectBinding", "unbindPod", "getBindings", "getBoundPods"],
        { params: true },
      )}
    />
  );
}

export function ChannelToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.channel`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      columns={threeColumnHeaders(t, prefix, "toolHeader", "descriptionHeader", "paramsHeader")}
      rows={buildDocsRows(
        t,
        prefix,
        [
          "searchChannels",
          "createChannel",
          "getChannel",
          "sendMessage",
          "getMessages",
          "getDocument",
          "updateDocument",
        ],
        { params: true },
      )}
    />
  );
}

export function TicketToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.ticket`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      columns={threeColumnHeaders(t, prefix, "toolHeader", "descriptionHeader", "paramsHeader")}
      rows={buildDocsRows(
        t,
        prefix,
        ["searchTickets", "getTicket", "createTicket", "updateTicket"],
        { params: true },
      )}
    />
  );
}

export function LoopToolsSection() {
  const t = useTranslations();
  const prefix = `${PREFIX}.loop`;
  return (
    <McpToolTableSection
      title={t(`${prefix}.title`)}
      description={t(`${prefix}.description`)}
      columns={threeColumnHeaders(t, prefix, "toolHeader", "descriptionHeader", "paramsHeader")}
      rows={buildDocsRows(t, prefix, ["listLoops", "triggerLoop"], { params: true })}
    />
  );
}
