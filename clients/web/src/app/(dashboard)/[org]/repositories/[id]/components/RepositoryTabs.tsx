"use client";

import { useTranslations } from "next-intl";
import { PillTabs, PillTabsRow } from "@/components/ui/pill-tabs";
import { RepositoryTab } from "./useRepositoryDetail";

interface RepositoryTabsProps {
  activeTab: RepositoryTab;
  onTabChange: (tab: RepositoryTab) => void;
}

export function RepositoryTabs({ activeTab, onTabChange }: RepositoryTabsProps) {
  const t = useTranslations();

  const tabs = [
    { id: "info", label: t("repositories.detail.information") },
    { id: "extensions", label: t("repositories.detail.extensions") },
  ];

  return (
    <PillTabsRow>
      <PillTabs active={activeTab} onChange={(id) => onTabChange(id as RepositoryTab)} tabs={tabs} />
    </PillTabsRow>
  );
}
