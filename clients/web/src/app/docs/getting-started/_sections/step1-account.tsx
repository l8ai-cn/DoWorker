"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export function Step1AccountSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={1} titleKey="docs.gettingStarted.step1.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step1.description")}</p>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm">
          <p className="font-medium mb-2">{t("docs.gettingStarted.step1.whatYouGet")}</p>
          <ul className="list-disc list-inside text-muted-foreground space-y-1">
            <li>{t("docs.gettingStarted.step1.item1")}</li>
            <li>{t("docs.gettingStarted.step1.item2")}</li>
          </ul>
        </div>
      </div>
    </section>
  );
}
