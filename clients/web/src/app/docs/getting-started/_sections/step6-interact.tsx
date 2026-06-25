"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export function Step6InteractSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={6} titleKey="docs.gettingStarted.step6.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step6.description")}</p>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          {Array.from({ length: 4 }, (_, i) => (
            <li key={i}>{t(`docs.gettingStarted.step6.item${i + 1}`)}</li>
          ))}
        </ul>
      </div>
    </section>
  );
}
