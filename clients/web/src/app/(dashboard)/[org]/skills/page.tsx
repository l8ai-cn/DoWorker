"use client";

import { useTranslations } from "next-intl";
import { SkillMarketPanel } from "@/components/settings/organization/extensions/SkillMarketPanel";

export default function SkillsMarketPage() {
  const t = useTranslations();

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-6xl">
        <SkillMarketPanel t={t} />
      </div>
    </div>
  );
}
