"use client";

import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useCurrentOrg } from "@/stores/auth";
import { useTranslations } from "next-intl";
import { Sparkles, Settings2, Store } from "lucide-react";

interface SkillsSidebarContentProps {
  className?: string;
}

export function SkillsSidebarContent({ className }: SkillsSidebarContentProps) {
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const t = useTranslations();
  const orgSlug = currentOrg?.slug ?? "";

  const goBrowse = () => {
    if (orgSlug) router.push(`/${orgSlug}/skills`);
  };

  const goRegistries = () => {
    if (orgSlug) {
      router.push(`/${orgSlug}/settings?scope=organization&tab=extensions`);
    }
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      <div className="p-3 border-b border-border">
        <h2 className="text-sm font-semibold">{t("ide.activities.skills")}</h2>
        <p className="text-xs text-muted-foreground mt-1">
          {t("ide.sidebar.skills.description")}
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-2 space-y-0.5">
        <SidebarLink icon={Store} label={t("ide.sidebar.skills.browse")} onClick={goBrowse} />
        <SidebarLink
          icon={Settings2}
          label={t("ide.sidebar.skills.manageRegistries")}
          onClick={goRegistries}
        />
      </div>
      <div className="bg-surface-muted/30 px-3 py-3 text-xs text-muted-foreground">
        <Sparkles className="w-3.5 h-3.5 inline mr-1.5" />
        {t("ide.sidebar.skills.hint")}
      </div>
    </div>
  );
}

function SidebarLink({
  icon: Icon,
  label,
  onClick,
}: {
  icon: typeof Store;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full flex items-center gap-2 px-3 py-2 text-sm rounded-md text-muted-foreground hover:bg-muted hover:text-foreground transition-colors text-left"
    >
      <Icon className="w-4 h-4 shrink-0" />
      <span className="truncate">{label}</span>
    </button>
  );
}

export default SkillsSidebarContent;
