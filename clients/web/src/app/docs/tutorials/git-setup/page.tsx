"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export default function GitSetupTutorialPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-2">
        {t("docs.tutorials.gitSetup.title")}
      </h1>
      <p className="text-sm text-muted-foreground mb-8">
        {t("docs.tutorials.gitSetup.difficulty")}
      </p>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.tutorials.gitSetup.description")}
      </p>

      {/* Prerequisites */}
      <section className="mb-8">
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-6">
          <h2 className="text-xl font-semibold mb-4 text-foreground">
            {t("docs.tutorials.gitSetup.prerequisites.title")}
          </h2>
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.prerequisites.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.gitSetup.prerequisites.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.prerequisites.item2")}</li>
          </ul>
        </div>
      </section>

      {/* Step 1: Open Git Settings */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={1} titleKey="docs.tutorials.gitSetup.step1.title" />
          <p className="text-muted-foreground">
            {t("docs.tutorials.gitSetup.step1.description")}
          </p>
        </div>
      </section>

      {/* Step 2: Connect GitHub */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={2} titleKey="docs.tutorials.gitSetup.step2.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step2.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.gitSetup.step2.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step2.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step2.item3")}</li>
            <li>{t("docs.tutorials.gitSetup.step2.item4")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.gitSetup.step2.tip")}
          </div>
        </div>
      </section>

      {/* Step 3: Connect GitLab */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={3} titleKey="docs.tutorials.gitSetup.step3.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step3.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.gitSetup.step3.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step3.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step3.item3")}</li>
            <li>{t("docs.tutorials.gitSetup.step3.item4")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.gitSetup.step3.tip")}
          </div>
        </div>
      </section>

      {/* Step 4: Connect Gitee */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={4} titleKey="docs.tutorials.gitSetup.step4.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step4.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.gitSetup.step4.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step4.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step4.item3")}</li>
          </ol>
        </div>
      </section>

      {/* Step 5: Import Repositories */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={5} titleKey="docs.tutorials.gitSetup.step5.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step5.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.gitSetup.step5.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step5.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step5.item3")}</li>
            <li>{t("docs.tutorials.gitSetup.step5.item4")}</li>
          </ol>
        </div>
      </section>

      {/* Step 6: SSH Keys */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={6} titleKey="docs.tutorials.gitSetup.step6.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step6.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.gitSetup.step6.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step6.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step6.item3")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.gitSetup.step6.tip")}
          </div>
        </div>
      </section>

      {/* Step 7: Verify */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={7} titleKey="docs.tutorials.gitSetup.step7.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.gitSetup.step7.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.gitSetup.step7.item1")}</li>
            <li>{t("docs.tutorials.gitSetup.step7.item2")}</li>
            <li>{t("docs.tutorials.gitSetup.step7.item3")}</li>
            <li>{t("docs.tutorials.gitSetup.step7.item4")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.gitSetup.step7.tip")}
          </div>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.gitSetup.nextSteps.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.tutorials.gitSetup.nextSteps.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Link
            href="/docs/tutorials/first-pod"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.gitSetup.nextSteps.item1")}
            </p>
          </Link>
          <Link
            href="/docs/tutorials/ticket-workflow"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.gitSetup.nextSteps.item2")}
            </p>
          </Link>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}
