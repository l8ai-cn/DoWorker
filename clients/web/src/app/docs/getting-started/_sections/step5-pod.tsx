"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";
import { LinkInText } from "../_components/link-in-text";

export function Step5PodSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={5} titleKey="docs.gettingStarted.step5.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step5.description")}</p>
        <ol className="list-decimal list-inside text-muted-foreground space-y-2">
          {Array.from({ length: 8 }, (_, i) => (
            <li key={i}>{t(`docs.gettingStarted.step5.item${i + 1}`)}</li>
          ))}
        </ol>
        <p className="text-sm text-muted-foreground mt-4">
          <LinkInText
            raw={t.raw("docs.gettingStarted.step5.seeSetup")}
            linkHref="/docs/tutorials/first-pod"
            linkLabel={t("docs.nav.tutorialFirstPod")}
          />
        </p>
      </div>
    </section>
  );
}
