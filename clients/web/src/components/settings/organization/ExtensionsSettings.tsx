"use client";

import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { SkillCatalogSettings, McpMarketSettings } from "./extensions";
import type { TranslationFn } from "./GeneralSettings";

interface ExtensionsSettingsProps {
  t: TranslationFn;
}

export function ExtensionsSettings({ t }: ExtensionsSettingsProps) {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold">{t("extensions.settings.title")}</h2>
        <p className="text-sm text-muted-foreground mt-1">
          {t("extensions.settings.description")}
        </p>
      </div>

      <Tabs defaultValue="skill-catalog" className="w-full">
        <TabsList>
          <TabsTrigger value="skill-catalog">
            {t("extensions.settings.tabs.skillCatalog")}
          </TabsTrigger>
          <TabsTrigger value="mcp-market">
            {t("extensions.settings.tabs.mcpMarket")}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="skill-catalog" className="mt-4">
          <SkillCatalogSettings t={t} />
        </TabsContent>

        <TabsContent value="mcp-market" className="mt-4">
          <McpMarketSettings t={t} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
