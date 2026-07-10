"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useCurrentUser, useCurrentOrg, useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import { LanguageSettings, ThemeSettings, NotificationSettings, AgentConfigPage, GitSettingsContent, AIResourcesSettings } from "@/components/settings";
import { GeneralSettings, MembersSettings, BillingSettings, APIKeysSettings, IMChannelsSettings, ExtensionsSettings, UsageSettings, InfrastructureOverview, ModelQuotasSettings } from "@/components/settings/organization";
import { SupportTicketsContent } from "@/components/support/SupportTicketsContent";
import { useTranslations } from "next-intl";
import { LogOut, User, Mail } from "lucide-react";

export default function SettingsPage() {
  const searchParams = useSearchParams();
  const scope = searchParams.get("scope") || "personal";
  const activeTab = searchParams.get("tab") || "general";
  const currentOrg = useCurrentOrg();
  const t = useTranslations();
  const translate = t as unknown as TranslationFn;

  const renderContent = () => {
    if (scope === "personal") {
      if (activeTab.startsWith("agents/")) {
        const agentSlug = activeTab.replace("agents/", "");
        return <AgentConfigPage agentSlug={agentSlug} />;
      }

      switch (activeTab) {
        case "general":
          return <PersonalGeneralSettings />;
        case "git":
          return <GitSettingsContent />;
        case "ai-resources":
          return <AIResourcesSettings scope="personal" canManage />;
        case "notifications":
          return <PersonalNotificationsSettings t={translate} />;
        case "support":
          return <SupportTicketsContent variant="narrow" />;
        default:
          return <PersonalGeneralSettings />;
      }
    }

    switch (activeTab) {
      case "general":
        return <GeneralSettings org={currentOrg} t={translate} />;
      case "members":
        return <MembersSettings t={translate} />;
      case "extensions":
        return <ExtensionsSettings t={translate} />;
      case "api-keys":
        return <APIKeysSettings t={translate} />;
      case "im-channels":
        return <IMChannelsSettings t={translate} />;
      case "billing":
        return <BillingSettings t={translate} />;
      case "usage":
        return <UsageSettings t={translate} />;
      case "ai-resources":
        return <AIResourcesSettings
          scope="organization"
          organizationSlug={currentOrg?.slug}
          canManage={currentOrg?.role === "owner" || currentOrg?.role === "admin"}
        />;
      case "model-quotas":
        return <ModelQuotasSettings />;
      case "infrastructure":
        return <InfrastructureOverview />;
      default:
        return <GeneralSettings org={currentOrg} t={translate} />;
    }
  };

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl">
        {renderContent()}
      </div>
    </div>
  );
}

function PersonalGeneralSettings() {
  const router = useRouter();
  const t = useTranslations();
  const user = useCurrentUser();
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <div className="space-y-6">
      {/* Account Information */}
      <div className="surface-card p-6">
        <h2 className="text-lg font-semibold mb-4">
          {t("settings.personal.general.accountInfo")}
        </h2>
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
              <User className="w-5 h-5 text-muted-foreground" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">
                {t("settings.personal.general.username")}
              </p>
              <p className="font-medium">{user?.username || "-"}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
              <Mail className="w-5 h-5 text-muted-foreground" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">
                {t("settings.personal.general.email")}
              </p>
              <p className="font-medium">{user?.email || "-"}</p>
            </div>
          </div>
        </div>
      </div>

      <LanguageSettings />
      <ThemeSettings />

      {/* Session / Logout */}
      <div className="surface-card p-6">
        <h2 className="text-lg font-semibold mb-2">
          {t("settings.personal.general.session")}
        </h2>
        <p className="text-sm text-muted-foreground mb-4">
          {t("settings.personal.general.sessionDescription")}
        </p>
        <Button
          variant="outline"
          onClick={handleLogout}
          className="flex items-center gap-2 text-destructive hover:text-destructive"
        >
          <LogOut className="w-4 h-4" />
          {t("settings.personal.general.logout")}
        </Button>
      </div>
    </div>
  );
}

type TranslationFn = (key: string, params?: Record<string, string | number>) => string;

function PersonalNotificationsSettings({ t }: { t: TranslationFn }) {
  return (
    <div className="space-y-6">
      <div className="surface-card p-6">
        <h2 className="text-lg font-semibold mb-4">{t("settings.notifications.title")}</h2>
        <p className="text-sm text-muted-foreground mb-6">
          {t("settings.notifications.description")}
        </p>
        <NotificationSettings />
      </div>
    </div>
  );
}
