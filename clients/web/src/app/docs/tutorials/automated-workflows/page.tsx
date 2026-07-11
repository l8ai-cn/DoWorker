"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export default function AutomatedWorkflowsTutorialPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-2">
        {t("docs.tutorials.workflows.title")}
      </h1>
      <p className="text-sm text-muted-foreground mb-8">
        {t("docs.tutorials.workflows.difficulty")}
      </p>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.tutorials.workflows.description")}
      </p>

      {/* What Are Workflows? */}
      <section className="mb-8">
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-6">
          <h2 className="text-xl font-semibold mb-4 text-foreground">
            {t("docs.tutorials.workflows.whatAreWorkflows.title")}
          </h2>
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.workflows.whatAreWorkflows.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.workflows.whatAreWorkflows.item1")}</li>
            <li>{t("docs.tutorials.workflows.whatAreWorkflows.item2")}</li>
            <li>{t("docs.tutorials.workflows.whatAreWorkflows.item3")}</li>
            <li>{t("docs.tutorials.workflows.whatAreWorkflows.item4")}</li>
          </ul>
        </div>
      </section>

      {/* Step 1 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={1} titleKey="docs.tutorials.workflows.step1.title" />
          <p className="text-muted-foreground">
            {t("docs.tutorials.workflows.step1.description")}
          </p>
        </div>
      </section>

      {/* Step 2 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={2} titleKey="docs.tutorials.workflows.step2.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.workflows.step2.description")}
          </p>
          <p className="font-medium mb-3">
            {t("docs.tutorials.workflows.step2.fields")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.workflows.step2.field1")}</li>
            <li>{t("docs.tutorials.workflows.step2.field2")}</li>
            <li>{t("docs.tutorials.workflows.step2.field3")}</li>
            <li>{t("docs.tutorials.workflows.step2.field4")}</li>
            <li>{t("docs.tutorials.workflows.step2.field5")}</li>
            <li>{t("docs.tutorials.workflows.step2.field6")}</li>
            <li>{t("docs.tutorials.workflows.step2.field7")}</li>
          </ol>
        </div>
      </section>

      {/* Step 3 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={3} titleKey="docs.tutorials.workflows.step3.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.workflows.step3.description")}
          </p>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto mb-4">
            <div className="space-y-2 text-success">
              <p>{t("docs.tutorials.workflows.step3.pattern1")}</p>
              <p>{t("docs.tutorials.workflows.step3.pattern2")}</p>
              <p>{t("docs.tutorials.workflows.step3.pattern3")}</p>
              <p>{t("docs.tutorials.workflows.step3.pattern4")}</p>
            </div>
          </div>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.workflows.step3.tip")}
          </div>
        </div>
      </section>

      {/* Step 4 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={4} titleKey="docs.tutorials.workflows.step4.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.workflows.step4.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.workflows.step4.item1")}</li>
            <li>{t("docs.tutorials.workflows.step4.item2")}</li>
            <li>{t("docs.tutorials.workflows.step4.item3")}</li>
            <li>{t("docs.tutorials.workflows.step4.item4")}</li>
          </ul>
        </div>
      </section>

      {/* Step 5 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={5} titleKey="docs.tutorials.workflows.step5.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.workflows.step5.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.workflows.step5.item1")}</li>
            <li>{t("docs.tutorials.workflows.step5.item2")}</li>
            <li>{t("docs.tutorials.workflows.step5.item3")}</li>
            <li>{t("docs.tutorials.workflows.step5.item4")}</li>
            <li>{t("docs.tutorials.workflows.step5.item5")}</li>
          </ul>
        </div>
      </section>

      {/* Common Patterns */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.workflows.commonPatterns.title")}
        </h2>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.tutorials.workflows.commonPatterns.pattern1.title")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.workflows.commonPatterns.pattern1.description")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.tutorials.workflows.commonPatterns.pattern2.title")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.workflows.commonPatterns.pattern2.description")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.tutorials.workflows.commonPatterns.pattern3.title")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.workflows.commonPatterns.pattern3.description")}
            </p>
          </div>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.workflows.nextSteps.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.tutorials.workflows.nextSteps.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Link
            href="/docs/tutorials/multi-agent-collaboration"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.workflows.nextSteps.item1")}
            </p>
          </Link>
          <Link
            href="/docs/tutorials/ticket-workflow"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.workflows.nextSteps.item2")}
            </p>
          </Link>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}
