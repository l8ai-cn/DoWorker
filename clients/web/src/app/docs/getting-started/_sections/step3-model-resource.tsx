"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export function Step3ModelResourceSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={3} titleKey="docs.gettingStarted.step3.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step3.description")}</p>
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 mt-4 text-sm text-muted-foreground">
          {t("docs.gettingStarted.step3.tip")}
        </div>
      </div>
    </section>
  );
}
