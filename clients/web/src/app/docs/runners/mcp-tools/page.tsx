"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import {
  OverviewSection,
  DiscoveryToolsSection,
  PodToolsSection,
  BindingToolsSection,
  ChannelToolsSection,
  TicketToolsSection,
  WorkflowToolsSection,
} from "./_sections/tool-sections";

export default function MCPToolsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">{t("docs.runners.mcpTools.title")}</h1>
      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.runners.mcpTools.description")}
      </p>

      <OverviewSection />
      <DiscoveryToolsSection />
      <PodToolsSection />
      <BindingToolsSection />
      <ChannelToolsSection />
      <TicketToolsSection />
      <WorkflowToolsSection />

      <DocNavigation />
    </div>
  );
}
