"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export default function MultiAgentTutorialPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-2">
        {t("docs.tutorials.multiAgent.title")}
      </h1>
      <p className="text-sm text-muted-foreground mb-8">
        {t("docs.tutorials.multiAgent.difficulty")}
      </p>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.tutorials.multiAgent.description")}
      </p>

      {/* Scenario Overview */}
      <section className="mb-8">
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-6">
          <h2 className="text-xl font-semibold mb-4 text-foreground">
            {t("docs.tutorials.multiAgent.scenario.title")}
          </h2>
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.scenario.description")}
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="surface-card p-4 bg-background">
              <h3 className="font-medium mb-1">
                {t("docs.tutorials.multiAgent.scenario.frontend").split(" — ")[0]}
              </h3>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.multiAgent.scenario.frontend").split(" — ")[1]}
              </p>
            </div>
            <div className="surface-card p-4 bg-background">
              <h3 className="font-medium mb-1">
                {t("docs.tutorials.multiAgent.scenario.backend").split(" — ")[0]}
              </h3>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.multiAgent.scenario.backend").split(" — ")[1]}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 1 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={1} titleKey="docs.tutorials.multiAgent.step1.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step1.description")}
          </p>
          <p className="font-medium mb-3">
            {t("docs.tutorials.multiAgent.step1.fields")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.multiAgent.step1.field1")}</li>
            <li>{t("docs.tutorials.multiAgent.step1.field2")}</li>
            <li>{t("docs.tutorials.multiAgent.step1.field3")}</li>
          </ol>
        </div>
      </section>

      {/* Step 2 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={2} titleKey="docs.tutorials.multiAgent.step2.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step2.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.multiAgent.step2.item1")}</li>
            <li>{t("docs.tutorials.multiAgent.step2.item2")}</li>
            <li>{t("docs.tutorials.multiAgent.step2.item3")}</li>
          </ul>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.multiAgent.step2.tip")}
          </div>
        </div>
      </section>

      {/* Step 3 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={3} titleKey="docs.tutorials.multiAgent.step3.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step3.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.multiAgent.step3.item1")}</li>
            <li>{t("docs.tutorials.multiAgent.step3.item2")}</li>
            <li>{t("docs.tutorials.multiAgent.step3.item3")}</li>
          </ul>
        </div>
      </section>

      {/* Step 4 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={4} titleKey="docs.tutorials.multiAgent.step4.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step4.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.multiAgent.step4.item1")}</li>
            <li>{t("docs.tutorials.multiAgent.step4.item2")}</li>
            <li>{t("docs.tutorials.multiAgent.step4.item3")}</li>
          </ul>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.multiAgent.step4.tip")}
          </div>
        </div>
      </section>

      {/* Step 5 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={5} titleKey="docs.tutorials.multiAgent.step5.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step5.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.multiAgent.step5.item1")}</li>
            <li>{t("docs.tutorials.multiAgent.step5.item2")}</li>
            <li>{t("docs.tutorials.multiAgent.step5.item3")}</li>
            <li>{t("docs.tutorials.multiAgent.step5.item4")}</li>
          </ul>
        </div>
      </section>

      {/* Step 6 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={6} titleKey="docs.tutorials.multiAgent.step6.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.multiAgent.step6.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.multiAgent.step6.item1")}</li>
            <li>{t("docs.tutorials.multiAgent.step6.item2")}</li>
            <li>{t("docs.tutorials.multiAgent.step6.item3")}</li>
            <li>{t("docs.tutorials.multiAgent.step6.item4")}</li>
          </ul>
        </div>
      </section>

      {/* Best Practices */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.multiAgent.bestPractices.title")}
        </h2>
        <div className="space-y-3">
          <div className="surface-card p-4">
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.multiAgent.bestPractices.tip1")}
            </p>
          </div>
          <div className="surface-card p-4">
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.multiAgent.bestPractices.tip2")}
            </p>
          </div>
          <div className="surface-card p-4">
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.multiAgent.bestPractices.tip3")}
            </p>
          </div>
          <div className="surface-card p-4">
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.multiAgent.bestPractices.tip4")}
            </p>
          </div>
          <div className="surface-card p-4">
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.multiAgent.bestPractices.tip5")}
            </p>
          </div>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}
